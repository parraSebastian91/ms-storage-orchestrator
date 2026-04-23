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
	m.logger.Info("Publishing message to RabbitMQ", map[string]interface{}{
		"routingKey":    routingKey,
		"correlationId": event.CorrelationId,
	})
	if exchange == "" {
		exchange = m.defaultExchange
	}
	var recipe interface{}
	var ok bool

	// La receta se selecciona SIEMPRE, independiente de si el routing key ya viene definido
	switch event.MediaType {
	case domainModels.MEDIA_TYPE_IMAGE:
		recipe, ok = domainModels.RECIPE_IMAGE[event.CategoryProcess]
		if !ok {
			recipe, ok = domainModels.RECIPE_IMAGE[domainModels.CATEGORY_PROCESS_USER_AVATAR]
		}
		if routingKey == "" {
			routingKey = MediaImageResize
		}
	case domainModels.MEDIA_TYPE_VIDEO:
		ok = true
		if routingKey == "" {
			routingKey = MediaVideoTranscode
		}
	case domainModels.MEDIA_TYPE_DOCUMENT:
		recipe, ok = domainModels.RECIPE_DOCUMENT[event.CategoryProcess]
		if !ok {
			recipe, ok = domainModels.RECIPE_DOCUMENT[domainModels.CATEGORY_PROCESS_DOCUMENT_DTO]
		}
		if routingKey == "" {
			routingKey = MediaDocumentUpload
		}
	default:
		ok = true
		if routingKey == "" {
			routingKey = MediaDocumentUpload
		}
	}

	headers := amqp.Table{
		"retry_count": 0,
		"requeue":     true,
	}

	if !ok {
		m.logger.Error("Recipe not found for category process", map[string]interface{}{
			"category":      event.CategoryProcess,
			"correlationId": event.CorrelationId,
		})
	}

	type publishPayload struct {
		Event         AplicationModel.StorageModel `json:"event"`
		Recipe        interface{}                  `json:"recipe"`
		CorrelationId string                       `json:"correlation_id"`
	}

	body, err := json.Marshal(publishPayload{
		Event:         event,
		Recipe:        recipe,
		CorrelationId: event.CorrelationId,
	})
	if err != nil {
		m.logger.Error("Failed to marshal publish payload", map[string]interface{}{
			"correlationId": event.CorrelationId,
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to marshal publish payload: %w", err)
	}

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
			Headers:      headers,
			Timestamp:    time.Now(),
		},
	)

	if err != nil {
		m.logger.Error("Failed to publish message", map[string]interface{}{
			"correlationId": event.CorrelationId,
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to publish message: %w", err)
	}
	m.logger.Info("Message published successfully", map[string]interface{}{
		"routingKey":    routingKey,
		"correlationId": event.CorrelationId,
	})
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
