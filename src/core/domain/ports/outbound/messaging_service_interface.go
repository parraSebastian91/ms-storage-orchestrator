package ports

import (
	"context"

	AplicationModel "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/model"
)

type IMessagePublisher interface {
	PublishTypeImage(ctx context.Context, message AplicationModel.StorageModel) error
	PublishTypeVideo(ctx context.Context, message AplicationModel.StorageModel) error
	PublishTypeDocument(ctx context.Context, message AplicationModel.StorageModel) error
	PublishTypeArchive(ctx context.Context, message AplicationModel.StorageModel) error
}

type IWorkPublisher interface {
	EnqueueWork(ctx context.Context, eventName string, payload []byte) error
}
