package application

import (
	"context"

	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/storageApplication/command"
)

type IStorageApplicationUseCase interface {
	UploadFile(ctx context.Context, uploadFormDTO command.UploadFileCommand) error
	DownloadFile(ctx context.Context, fileName string) ([]byte, error)
	DeleteFile(ctx context.Context, fileName string) error
	ListFiles(ctx context.Context) ([]string, error)
}
