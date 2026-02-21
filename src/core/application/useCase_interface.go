package application

import (
	"context"
	"io"
)

type IStorageApplicationUseCase interface {
	UploadFile(ctx context.Context, fileData io.Reader, fileName string, size int64) error
	DownloadFile(ctx context.Context, fileName string) ([]byte, error)
	DeleteFile(ctx context.Context, fileName string) error
	ListFiles(ctx context.Context) ([]string, error)
}
