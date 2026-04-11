package messaging

import (
	"context"

	AplicationModel "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/model"
	ports "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
)

var _ ports.IMessagePublisher = (*MessagingPublisherImpl)(nil)

type MessagingPublisherImpl struct {
	QueueClient *MessagingPublisherClient
	logger      ports.ILoggerService
}

func NewMessagingPublisherImpl(queueClient *MessagingPublisherClient, logger ports.ILoggerService) *MessagingPublisherImpl {
	return &MessagingPublisherImpl{
		QueueClient: queueClient,
		logger:      logger,
	}
}

func (m *MessagingPublisherImpl) PublishTypeImage(ctx context.Context, event AplicationModel.StorageModel) error {
	m.QueueClient.Publish(ctx, m.QueueClient.defaultExchange, m.QueueClient.defaultExchange, []byte(event.StorageKey))
	return nil
}

func (m *MessagingPublisherImpl) PublishTypeVideo(ctx context.Context, event AplicationModel.StorageModel) error {
	// Implementa la lógica para publicar el mensaje utilizando msgPublishClient
	return nil
}

func (m *MessagingPublisherImpl) PublishTypeDocument(ctx context.Context, event AplicationModel.StorageModel) error {
	// Implementa la lógica para publicar el mensaje utilizando msgPublishClient
	return nil
}

func (m *MessagingPublisherImpl) PublishTypeArchive(ctx context.Context, event AplicationModel.StorageModel) error {
	// Implementa la lógica para publicar el mensaje utilizando msgPublishClient
	return nil
}
