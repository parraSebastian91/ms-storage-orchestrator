package ports

import (
	"context"

	AplicationModel "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/model"
)

type IMediaRepository interface {
	CreateMediaMetadata(ctx context.Context, model AplicationModel.StorageModel) error
	GetMediaMetadata(ctx context.Context, objkectKey string) (AplicationModel.StorageModel, error)
}
