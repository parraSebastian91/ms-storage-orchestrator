package ports

import (
	"context"

	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/useCase/storageApplication/command"
)

type IStorageUseCase interface {
	UploadFile(ctx context.Context, uploadFormDTO command.UploadFileCommand) error
	ExecuteProcessFile(ctx context.Context, objectKey string) error
	ExecuteGetPresignedPutURL(ctx context.Context, command command.GetPresignedPutURLCommand) (string, error)
}
