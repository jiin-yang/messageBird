package message

import (
	"context"
	"github.com/jiin-yang/messageBird/internal/client/webhook"
	"github.com/jiin-yang/messageBird/internal/infra/rabbitmq"
	"github.com/rs/zerolog/log"
	"sync"
)

type UseCase interface {
	CreateMessage(ctx context.Context, requestMsg CreateMessageRequest) (*CreateMessageResponse, error)
	GetOldestStatusNewMessages(ctx context.Context) ([]Message, error)
	SendMessages(ctx context.Context) error
	GetSentStatusMessages(ctx context.Context) ([]GetMessageResponse, error)
	StartConsumeFailures(ctx context.Context, maxRetries int)
	StopConsumeFailures()
}

type useCase struct {
	repo     Repository
	webhook  webhook.Client
	rabbitMQ rabbitmq.Client

	mu                sync.Mutex
	isConsumerRunning bool
	consumerCancel    context.CancelFunc
}

type NewUseCaseOptions struct {
	Repo     Repository
	Webhook  webhook.Client
	RabbitMQ rabbitmq.Client
}

func NewUseCase(opts *NewUseCaseOptions) UseCase {
	return &useCase{
		repo:     opts.Repo,
		webhook:  opts.Webhook,
		rabbitMQ: opts.RabbitMQ,
	}
}

func (u *useCase) CreateMessage(ctx context.Context, requestMsg CreateMessageRequest) (*CreateMessageResponse, error) {
	msg := CreateMessage{
		PhoneNumber: requestMsg.PhoneNumber,
		Content:     requestMsg.Content,
		Status:      New,
	}

	dbRes, err := u.repo.CreateMessage(ctx, msg)
	if err != nil {
		return nil, err
	}

	createdMsgRes := CreateMessageResponse{
		Id:          dbRes.Id,
		PhoneNumber: dbRes.PhoneNumber,
		Content:     dbRes.Content,
		Status:      dbRes.Status.String(),
		CreatedAt:   dbRes.CreatedAt,
	}

	return &createdMsgRes, err
}

func (u *useCase) GetOldestStatusNewMessages(ctx context.Context) ([]Message, error) {
	messages, err := u.repo.GetOldestStatusNewMessages(ctx)
	if err != nil {
		return nil, err
	}
	return messages, err
}

func (u *useCase) SendMessages(ctx context.Context) error {
	messages, err := u.GetOldestStatusNewMessages(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch oldest messages with status 'new'")
		return err
	}

	sendMsg := webhook.SendMessageRequest{}
	for _, message := range messages {
		sendMsg.To = message.PhoneNumber
		sendMsg.Content = message.Content

		err = u.repo.UpdateMessageStatus(ctx, message.Id, Process)
		if err != nil {
			log.Error().Err(err).Str("messageId", message.Id).Msg("Failed to update message status to 'Process'")
			continue
		}

		respWebhook, err := u.webhook.SendMessage(sendMsg)
		if err != nil {
			log.Error().Err(err).Str("messageId", message.Id).Msg("Failed to send message to webhook")

			err = u.repo.UpdateMessageStatus(ctx, message.Id, Fail)
			if err != nil {
				log.Error().Err(err).Str("messageId", message.Id).Msg("Failed to update message status to 'Fail'")
			}

			failedMsg := rabbitmq.FailedMessage{
				MessageID:   message.Id,
				PhoneNumber: message.PhoneNumber,
				Content:     message.Content,
				Status:      uint8(Fail),
			}
			pubErr := u.rabbitMQ.PublishFailMessage(ctx, failedMsg)
			if pubErr != nil {
				log.Error().Err(pubErr).
					Str("messageId", message.Id).
					Msg("Failed to publish fail message to RabbitMQ")
			}

			continue
		}

		log.Info().Msgf("Webhook response: %v %v", respWebhook.ResponseId, respWebhook.State)

		err = u.repo.UpdateMessageStatus(ctx, message.Id, Sent)
		if err != nil {
			// Message gonderildi fakat statu process->sent islemi yapilamadi. Bu durum simdilik Allah'a emanet
			// Message'i webhook.siteye gonderdigim icin kuyruga da atamiyorum tekrardan
			log.Error().Err(err).Str("messageId", message.Id).Msg("Failed to update message status to 'Sent'")
			continue
		}
	}
	return nil
}

func (u *useCase) GetSentStatusMessages(ctx context.Context) ([]GetMessageResponse, error) {
	messages, err := u.repo.GetSentStatusMessages(ctx)
	if err != nil {
		return nil, err
	}
	respMessages := []GetMessageResponse{}
	var respMsg GetMessageResponse
	for _, msg := range messages {
		respMsg.Id = msg.Id
		respMsg.Content = msg.Content
		respMsg.Status = msg.Status.String()
		respMsg.PhoneNumber = msg.PhoneNumber
		respMsg.CreatedAt = msg.CreatedAt
		respMsg.UpdatedAt = msg.UpdatedAt

		respMessages = append(respMessages, respMsg)
	}

	return respMessages, nil
}

func (u *useCase) StartConsumeFailures(ctx context.Context, maxRetries int) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.isConsumerRunning {
		log.Warn().Msg("RabbitMQ consumer is already running")
		return
	}

	log.Info().Msg("Starting RabbitMQ consumer")
	u.isConsumerRunning = true

	consumerCtx, cancel := context.WithCancel(ctx)
	u.consumerCancel = cancel

	go func() {
		defer func() {
			u.mu.Lock()
			u.isConsumerRunning = false
			u.mu.Unlock()
			log.Info().Msg("RabbitMQ consumer has stopped")
		}()

		retryTask := func(msg rabbitmq.FailedMessage) error {
			log.Info().
				Str("messageId", msg.MessageID).
				Int("attempt", msg.Attempt).
				Msg("Retrying failed message")

			req := webhook.SendMessageRequest{
				To:      msg.PhoneNumber,
				Content: msg.Content,
			}
			resp, err := u.webhook.SendMessage(req)
			if err != nil {
				return err
			}

			log.Info().Msgf("Retry webhook response: %v %v", resp.ResponseId, resp.State)
			return nil
		}

		updateStatus := func(messageID string, status uint8) error {
			return u.repo.UpdateMessageStatus(ctx, messageID, Status(status))
		}

		err := u.rabbitMQ.ConsumeFailures(consumerCtx, retryTask, updateStatus, maxRetries)
		if err != nil {
			log.Error().Err(err).Msg("RabbitMQ consumer encountered an error")
		}
	}()
}

func (u *useCase) StopConsumeFailures() {
	u.mu.Lock()
	defer u.mu.Unlock()

	if !u.isConsumerRunning {
		log.Warn().Msg("RabbitMQ consumer is not running")
		return
	}

	log.Info().Msg("Stopping RabbitMQ consumer")
	u.consumerCancel()
	u.isConsumerRunning = false
}
