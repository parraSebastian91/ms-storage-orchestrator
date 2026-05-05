package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	AplicationModel "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/model"
	domainModels "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/models"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/config"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/infrastructure/observability"
	amqp "github.com/rabbitmq/amqp091-go"
)

type MessagingPublisherClient struct {
	url         string
	connection  *amqp.Connection
	channel     *amqp.Channel
	Exchanges   config.ExchangeConfig
	RoutingKeys config.RoutingKeysConfig
	logger      *observability.CustomLogger
	mu          sync.Mutex
}

func NewMessagingPublisherClient(url string, exchanges config.ExchangeConfig, routingKeys config.RoutingKeysConfig, logger *observability.CustomLogger) (*MessagingPublisherClient, error) {

	connection, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to RabbitMQ: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		connection.Close()
		return nil, fmt.Errorf("Failed to open a channel: %w", err)
	}

	if err := declareExchanges(channel, exchanges); err != nil {
		channel.Close()
		connection.Close()
		return nil, fmt.Errorf("failed to declare RabbitMQ exchanges: %w", err)
	}

	logger.Info("Messaging publisher initialized successfully", nil)

	return &MessagingPublisherClient{
		url:         url,
		connection:  connection,
		channel:     channel,
		Exchanges:   exchanges,
		RoutingKeys: routingKeys,
		logger:      logger,
	}, nil
}

func declareExchanges(channel *amqp.Channel, exchanges config.ExchangeConfig) error {
	if err := channel.ExchangeDeclare(
		exchanges.ExchangeTask,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare task exchange %q: %w", exchanges.ExchangeTask, err)
	}

	if err := channel.ExchangeDeclare(
		exchanges.ExchangeNotification,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return fmt.Errorf("failed to declare notification exchange %q: %w", exchanges.ExchangeNotification, err)
	}

	return nil
}

// ensureConnection checks if the connection is still open and reconnects if needed
func (m *MessagingPublisherClient) ensureConnection(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connection == nil || m.channel == nil || m.connection.IsClosed() || m.channel.IsClosed() {
		m.logger.Warn("RabbitMQ connection is closed, attempting to reconnect", map[string]interface{}{
			"url": m.url,
		})

		// Close old connection if exists
		if m.channel != nil {
			m.channel.Close()
		}
		if m.connection != nil {
			m.connection.Close()
		}

		// Attempt reconnection with exponential backoff
		var lastErr error
		maxRetries := 3
		for attempt := 0; attempt < maxRetries; attempt++ {
			backoff := time.Duration(1<<uint(attempt)) * time.Second // 1s, 2s, 4s

			if attempt > 0 {
				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			connection, err := amqp.Dial(m.url)
			if err != nil {
				lastErr = err
				m.logger.Warn("Failed to reconnect to RabbitMQ", map[string]interface{}{
					"attempt": attempt + 1,
					"error":   err.Error(),
				})
				continue
			}

			channel, err := connection.Channel()
			if err != nil {
				connection.Close()
				lastErr = err
				m.logger.Warn("Failed to open channel after reconnection", map[string]interface{}{
					"attempt": attempt + 1,
					"error":   err.Error(),
				})
				continue
			}

			if err := declareExchanges(channel, m.Exchanges); err != nil {
				channel.Close()
				connection.Close()
				lastErr = err
				m.logger.Warn("Failed to declare exchanges after reconnection", map[string]interface{}{
					"attempt": attempt + 1,
					"error":   err.Error(),
				})
				continue
			}

			m.connection = connection
			m.channel = channel
			m.logger.Info("Successfully reconnected to RabbitMQ", nil)
			return nil
		}

		return fmt.Errorf("failed to reconnect to RabbitMQ after %d attempts: %w", maxRetries, lastErr)
	}

	return nil
}

func (m *MessagingPublisherClient) reconnectNow(ctx context.Context) error {
	m.mu.Lock()
	if m.channel != nil {
		_ = m.channel.Close()
	}
	if m.connection != nil {
		_ = m.connection.Close()
	}
	m.channel = nil
	m.connection = nil
	m.mu.Unlock()

	return m.ensureConnection(ctx)
}

func isChannelNotOpenErr(err error) bool {
	if err == nil {
		return false
	}

	if amqpErr, ok := err.(*amqp.Error); ok && amqpErr.Code == 504 {
		return true
	}

	return strings.Contains(err.Error(), "channel/connection is not open")
}

func (m *MessagingPublisherClient) publishWithRetry(
	ctx context.Context,
	exchange string,
	routingKey string,
	message amqp.Publishing,
	correlationID string,
	logLabel string,
) error {
	// First attempt
	m.mu.Lock()
	channel := m.channel
	m.mu.Unlock()

	err := channel.PublishWithContext(ctx, exchange, routingKey, false, false, message)
	if err == nil {
		return nil
	}

	m.logger.Warn("Publish failed, trying to recover RabbitMQ channel/connection", map[string]interface{}{
		"correlationId": correlationID,
		"error":         err.Error(),
		"exchange":      exchange,
		"routingKey":    routingKey,
		"label":         logLabel,
	})

	var reconnectErr error
	if isChannelNotOpenErr(err) {
		reconnectErr = m.reconnectNow(ctx)
	} else {
		reconnectErr = m.ensureConnection(ctx)
	}

	if reconnectErr != nil {
		return fmt.Errorf("publish failed and reconnect failed: %w (publish error: %v)", reconnectErr, err)
	}

	// Retry once with a fresh channel/connection
	m.mu.Lock()
	channel = m.channel
	m.mu.Unlock()

	err = channel.PublishWithContext(ctx, exchange, routingKey, false, false, message)
	if err != nil {
		return fmt.Errorf("publish failed after reconnect retry: %w", err)
	}

	m.logger.Info("Publish succeeded after RabbitMQ reconnect retry", map[string]interface{}{
		"correlationId": correlationID,
		"exchange":      exchange,
		"routingKey":    routingKey,
		"label":         logLabel,
	})

	return nil
}

func (m *MessagingPublisherClient) PublishTask(ctx context.Context, routingKey string, event AplicationModel.StorageModel) error {
	// Ensure connection is still open
	if err := m.ensureConnection(ctx); err != nil {
		m.logger.Error("Failed to ensure RabbitMQ connection", map[string]interface{}{
			"correlationId": event.CorrelationId,
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to ensure RabbitMQ connection: %w", err)
	}

	exchange := m.Exchanges.ExchangeTask

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
			routingKey = m.RoutingKeys.MediaImageResize
		}
	case domainModels.MEDIA_TYPE_VIDEO:
		ok = true
		if routingKey == "" {
			routingKey = m.RoutingKeys.MediaVideoTranscode
		}
	case domainModels.MEDIA_TYPE_DOCUMENT:
		recipe, ok = domainModels.RECIPE_DOCUMENT[event.CategoryProcess]
		if !ok {
			recipe, ok = domainModels.RECIPE_DOCUMENT[domainModels.CATEGORY_PROCESS_DOCUMENT_DTO]
		}
		if routingKey == "" {
			routingKey = m.RoutingKeys.MediaDocumentUpload
		}
	default:
		ok = true
		if routingKey == "" {
			routingKey = m.RoutingKeys.DteProcessNotification
		}
	}

	m.logger.Info("Publishing message to RabbitMQ", map[string]interface{}{
		"exchange":      exchange,
		"routingKey":    routingKey,
		"correlationId": event.CorrelationId,
	})

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

	err = m.publishWithRetry(
		ctx,
		exchange,
		routingKey,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Headers:      headers,
			Timestamp:    time.Now(),
		},
		event.CorrelationId,
		"task",
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

func (m *MessagingPublisherClient) PublishNotification(ctx context.Context, event AplicationModel.NotifyModel) error {
	// Ensure connection is still open
	if err := m.ensureConnection(ctx); err != nil {
		m.logger.Error("Failed to ensure RabbitMQ connection", map[string]interface{}{
			"correlationId": event.CorrelationId,
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to ensure RabbitMQ connection: %w", err)
	}

	m.logger.Info("Publishing notification message to RabbitMQ", map[string]interface{}{
		"routingKey":    m.RoutingKeys.DteProcessNotification,
		"correlationId": event.CorrelationId,
	})
	exchange := m.Exchanges.ExchangeNotification
	routingKey := m.RoutingKeys.DteProcessNotification

	type publishPayload struct {
		Event         AplicationModel.NotifyModel `json:"event"`
		CorrelationId string                      `json:"correlation_id"`
	}

	body, err := json.Marshal(publishPayload{
		Event:         event,
		CorrelationId: event.CorrelationId,
	})
	if err != nil {
		m.logger.Error("Failed to marshal publish payload", map[string]interface{}{
			"correlationId": event.CorrelationId,
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to marshal publish payload: %w", err)
	}

	err = m.publishWithRetry(
		ctx,
		exchange,
		routingKey,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
			Timestamp:    time.Now(),
		},
		event.CorrelationId,
		"notification",
	)

	if err != nil {
		m.logger.Error("Failed to publish notification message", map[string]interface{}{
			"correlationId": event.CorrelationId,
			"error":         err.Error(),
		})
		return fmt.Errorf("failed to publish notification message: %w", err)
	}
	m.logger.Info("Notification message published successfully", map[string]interface{}{
		"routingKey":    routingKey,
		"correlationId": event.CorrelationId,
	})
	return nil
}

func (m *MessagingPublisherClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.channel != nil {
		m.channel.Close()
	}
	if m.connection != nil {
		m.connection.Close()
	}
	return nil
}
