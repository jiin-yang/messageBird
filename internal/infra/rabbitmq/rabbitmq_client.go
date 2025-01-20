package rabbitmq

import (
	"context"
	"encoding/json"
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"time"
)

type Status uint8

const (
	New Status = iota + 1
	Process
	Sent
	Fail
	Dead
)

type Client interface {
	PublishFailMessage(ctx context.Context, msg FailedMessage) error
	Close() error
	ConsumeFailures(ctx context.Context,
		retryTask func(FailedMessage) error,
		updateStatus func(messageID string, status uint8) error, maxRetries int) error
}

type FailedMessage struct {
	MessageID   string `json:"messageId"`
	PhoneNumber string `json:"phoneNumber"`
	Content     string `json:"content"`
	Attempt     int    `json:"attempt"`
	Status      uint8  `json:"status"`
}

type client struct {
	conn          *amqp091.Connection
	channel       *amqp091.Channel
	failQueueName string
}

func NewRabbitMQClient(amqpURL, failQueueName string) (Client, error) {
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to RabbitMQ via amqp091-go")
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Error().Err(err).Msg("Failed to open a channel in RabbitMQ")
		return nil, err
	}

	_, err = ch.QueueDeclare(failQueueName, true, false, false, false, nil)
	if err != nil {
		log.Error().Err(err).Str("queue", failQueueName).Msg("Failed to declare fail_messages queue")
		return nil, err
	}

	return &client{
		conn:          conn,
		channel:       ch,
		failQueueName: failQueueName,
	}, nil
}

func (c *client) PublishFailMessage(ctx context.Context, msg FailedMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		log.Error().
			Err(err).
			Interface("failedMessage", msg).
			Msg("Failed to marshal FailedMessage")
		return err
	}

	err = c.channel.PublishWithContext(
		ctx,
		"",
		c.failQueueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("queue", c.failQueueName).
			Msg("Failed to publish message to RabbitMQ")
		return err
	}

	log.Info().
		Str("queue", c.failQueueName).
		Str("messageId", msg.MessageID).
		Msg("Fail message published to RabbitMQ successfully")

	return nil
}

func (c *client) Close() error {
	if err := c.channel.Close(); err != nil {
		log.Warn().Err(err).Msg("Failed to close RabbitMQ channel")
	}

	if err := c.conn.Close(); err != nil {
		log.Warn().Err(err).Msg("Failed to close RabbitMQ connection")
	}
	return nil
}

func (c *client) ConsumeFailures(
	ctx context.Context,
	retryTask func(FailedMessage) error,
	updateStatus func(messageID string, status uint8) error,
	maxRetries int,
) error {
	messages, err := c.channel.Consume(
		c.failQueueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start consuming from fail_messages queue")
		return err
	}

	for delivery := range messages {
		select {
		case <-ctx.Done():
			log.Warn().Msg("Consumer stopped by context cancellation")
			return nil
		default:
		}

		var failMsg FailedMessage
		if err = json.Unmarshal(delivery.Body, &failMsg); err != nil {
			log.Error().Err(err).Msg("Failed to parse message")
			if err = updateStatus(failMsg.MessageID, uint8(Dead)); err != nil {
				log.Error().Err(err).Msg("Failed to update message status to Dead - rabbitmq-unmarshall")
			}
			_ = delivery.Ack(false)
			continue
		}

		log.Info().Msgf("MessageId: %v, Content: %v, Attempt: %d", failMsg.MessageID, failMsg.Content, failMsg.Attempt)

		if failMsg.Attempt >= maxRetries {
			log.Warn().Str("messageId", failMsg.MessageID).Msg("Max retries reached, discarding message")
			if err = updateStatus(failMsg.MessageID, uint8(Dead)); err != nil {
				log.Error().Err(err).Msg("Failed to update message status to Dead - rabbitmq-maxRetries")
			}
			_ = delivery.Ack(false)
			continue
		}

		if err = retryTask(failMsg); err != nil {
			log.Error().Err(err).Str("messageId", failMsg.MessageID).Msg("Failed to process message, retrying")
			failMsg.Attempt++
			if failMsg.Attempt == 1 {
				err = updateStatus(failMsg.MessageID, uint8(Fail))
				log.Error().Err(err).Msg("Failed to update message status to Fail - rabbitmq-retryTask")
			}
			time.Sleep(time.Duration(fibonacci(failMsg.Attempt)) * time.Second)

			if err = c.PublishFailMessage(ctx, failMsg); err != nil {
				log.Error().Err(err).Str("messageId", failMsg.MessageID).Msg("Failed to republish message for retry")
			}
			_ = delivery.Ack(false)

		} else {
			log.Info().Str("messageId", failMsg.MessageID).Msg("Message sent successfully")
			if err = updateStatus(failMsg.MessageID, uint8(Sent)); err != nil {
				log.Error().Err(err).Msg("Failed to update message status to Sent - rabbitmq-retryTask")
			}
			_ = delivery.Ack(false)
		}
	}

	return nil
}

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}
