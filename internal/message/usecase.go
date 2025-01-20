package message

import (
	"context"
	"github.com/jiin-yang/messageBird/internal/client/webhook"
	"github.com/rs/zerolog/log"
)

type UseCase interface {
	CreateMessage(ctx context.Context, requestMsg CreateMessageRequest) (*CreateMessageResponse, error)
	GetOldestStatusNewMessages(ctx context.Context) ([]Message, error)
	SendMessages(ctx context.Context) error
}

type useCase struct {
	repo    Repository
	webhook webhook.Client
}

type NewUseCaseOptions struct {
	Repo    Repository
	Webhook webhook.Client
}

func NewUseCase(opts *NewUseCaseOptions) UseCase {
	return &useCase{
		repo:    opts.Repo,
		webhook: opts.Webhook,
	}
}

func (u useCase) CreateMessage(ctx context.Context, requestMsg CreateMessageRequest) (*CreateMessageResponse, error) {
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

func (u useCase) GetOldestStatusNewMessages(ctx context.Context) ([]Message, error) {
	messages, err := u.repo.GetOldestStatusNewMessages(ctx)
	if err != nil {
		return nil, err
	}
	return messages, err
}

func (u useCase) SendMessages(ctx context.Context) error {
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
			// Message gonderilemedi ve statusu processte
			// Queue - Retry kuyruğuna ekleme yapilacak
			continue
		}

		log.Info().Msgf("Webhook response: %v %v", respWebhook.ResponseId, respWebhook.State)

		err = u.repo.UpdateMessageStatus(ctx, message.Id, Sent)
		if err != nil {
			// Message gonderildi fakat statu process->sent islemi yapilamadi. Bu durum simdilik Allah'a emanet
			log.Error().Err(err).Str("messageId", message.Id).Msg("Failed to update message status to 'Sent'")
			continue
		}
	}
	return nil
}
