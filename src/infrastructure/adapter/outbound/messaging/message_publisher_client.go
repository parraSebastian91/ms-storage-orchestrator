package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	AplicationModel "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/model"
	domainModels "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/models"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/observability"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	MediaImageResize    string = "media.image.resize"
	MediaVideoTranscode string = "media.video.transcode"
	MediaDocumentUpload string = "media.document.upload"
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
		MediaImageResize,
		MediaVideoTranscode,
		MediaDocumentUpload,
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
		logger:          logger,
	}, nil
}

func (m *MessagingPublisherClient) Publish(ctx context.Context, exchange string, routingKey string, event AplicationModel.StorageModel) error {
	if exchange == "" {
		exchange = m.defaultExchange
	}

	if routingKey == "" {
		switch event.MediaType {
		case domainModels.MEDIA_TYPE_IMAGE:
			routingKey = MediaImageResize
		case domainModels.MEDIA_TYPE_VIDEO:
			routingKey = MediaVideoTranscode
		case domainModels.MEDIA_TYPE_DOCUMENT:
			routingKey = MediaDocumentUpload
		default:
			routingKey = MediaDocumentUpload
		}
	}

	// headers := amqp.Table{
	// 	"AssetId":       event.AssetId,
	// 	"owner_id":      event.OwnerUUID,
	// 	"media_type":    event.MediaType,
	// 	"category":      event.CategoryProcess,
	// 	"original_name": event.NameFile,
	// 	"storage_key":   event.StorageKey,
	// }

	recipe, ok := domainModels.RECIPE[event.CategoryProcess]
	if !ok {
		m.logger.Error("Recipe not found for category process", map[string]interface{}{
			"category": event.CategoryProcess,
		})
	}

	type publishPayload struct {
		Event  AplicationModel.StorageModel  `json:"event"`
		Recipe domainModels.RecipeMediaModel `json:"recipe"`
	}

	body, err := json.Marshal(publishPayload{
		Event:  event,
		Recipe: recipe,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal publish payload: %w", err)
	}

	fmt.Printf("Publishing message with routing key: %s\n", routingKey)
	fmt.Printf("Publishing message with queue: %s\n", m.defaultQueue)
	fmt.Printf("Publishing message with exchange: %s\n", exchange)

	err = m.channel.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			// Headers:      headers,
			Timestamp: time.Now(),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (m *MessagingPublisherClient) Close() error {
	if m.channel != nil {
		m.channel.Close()
	}
	if m.connection != nil {
		m.connection.Close()
	}
	return nil
}
