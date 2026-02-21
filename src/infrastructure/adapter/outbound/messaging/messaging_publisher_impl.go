package messaging

import (
	"context"

	ports "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
)

var _ ports.IMessagePublisher = (*MessagingPublisherImpl)(nil)
var _ ports.IWorkPublisher = (*MessagingPublisherImpl)(nil)

type MessagingPublisherImpl struct {
	msgPublishClient *MessagingPublisherClient
	logger           ports.ILoggerService
}

func NewMessagingPublisherImpl(msgPublishClient *MessagingPublisherClient, logger ports.ILoggerService) *MessagingPublisherImpl {
	return &MessagingPublisherImpl{
		msgPublishClient: msgPublishClient,
		logger:           logger,
	}
}

func (m *MessagingPublisherImpl) Publish(ctx context.Context, exchange string, message []byte) error {

	return nil
}

func (m *MessagingPublisherImpl) PublishJson(ctx context.Context, exchange string, message interface{}) error {
	// Implementa la lógica para convertir el mensaje a JSON y publicarlo en RabbitMQ
	return nil
}

func (m *MessagingPublisherImpl) EnqueueWork(ctx context.Context, eventName string, payload []byte) error {
	// Implementa la lógica para encolar el trabajo en RabbitMQ
	return nil
}
