package messaging

import (
	"context"
	"fmt"
	"time"

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

func (m *MessagingPublisherClient) Publish(ctx context.Context, exchange string, routingKey string, message []byte) error {
	if exchange == "" {
		exchange = m.defaultExchange
	}

	headers := amqp.Table{
		"trace_id":    data.Header.TraceId,
		"code":        data.Header.Code,
		"name":        data.Header.Name,
		"file_mime":   data.Header.FileMime,
		"retry_count": data.Header.RetryCount,
		"chunked":     data.Header.Chunked,
		"chunk_index": data.Header.ChunkIndex,
		"is_last":     data.Header.IsLast,
		"backup":      data.Header.Backup,
	}

	fmt.Printf("Publishing message with headers: %v and routing key: %s\n", headers, routingKey)

	err := r.channel.PublishWithContext(
		ctx,
		exchange,        // exchange
		"upload.object", // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // 2 = persistente
			ContentType:  "application/octet-stream",
			Body:         data.File.File,
			Headers:      headers,
			Timestamp:    time.Now(),
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
