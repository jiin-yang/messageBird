package rabbitmq

import (
	"github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type Client interface {
	PublishFailMessage(msg FailedMessage) error
	Close() error
}

type FailedMessage struct {
	MessageID   string `json:"messageId"`
	PhoneNumber string `json:"phoneNumber"`
	Content     string `json:"content"`
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

func (c client) PublishFailMessage(msg FailedMessage) error {
	//TODO implement me
	panic("implement me")
}

func (c client) Close() error {
	//TODO implement me
	panic("implement me")
}
