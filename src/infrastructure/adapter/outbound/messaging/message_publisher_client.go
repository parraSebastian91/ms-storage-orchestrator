package messaging

import (
	"fmt"

	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/observability"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MessagingPublisherClient struct {
	connection      *amqp.Connection
	channel         *amqp.Channel
	defaultExchange string
	defaultQueue    string
	logger          *observability.CustomLogger
}

func NewMessagingPublisherClient(url string, defaultExchange string, defaultQueue string, logger *observability.CustomLogger) (*MessagingPublisherClient, error) {

	connection, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		return nil, fmt.Errorf("Failed to open a channel: %w", err)
	}

	err = channel.ExchangeDeclare(
		defaultExchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		channel.Close()
		connection.Close()
		return nil, fmt.Errorf("Failed to declare exchange: %w", err)
	}

	_, err = channel.QueueDeclare(
		defaultQueue,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		channel.Close()
		connection.Close()
		return nil, fmt.Errorf("Failed to declare queue: %w", err)
	}

	routingKey := []string{
		"media.image.resize",
		"media.video.transcode",
		"media.document.convert",
	}

	for _, key := range routingKey {
		err = channel.QueueBind(
			defaultQueue,
			key,
			defaultExchange,
			false,
			nil,
		)
		if err != nil {
			channel.Close()
			connection.Close()
			return nil, fmt.Errorf("Failed to bind queue: %w", err)
		}
	}
	logger.Info("Messaging publisher initialized successfully", nil)

	return &MessagingPublisherClient{
		connection:      connection,
		channel:         channel,
		defaultExchange: defaultExchange,
		defaultQueue:    defaultQueue,
	}, nil
}

func (m *MessagingPublisherClient) Close() error {
	// Implementa la lógica para cerrar cualquier conexión o recurso utilizado por el publisher
	return nil
}
