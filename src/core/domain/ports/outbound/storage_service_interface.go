package ports

import (
	"context"
	"io"
)

type IStorageService interface {
	DownloadFile(ctx context.Context, fileName string) (io.Reader, error)
	UploadFile(ctx context.Context, fileName string, fileContent io.Reader, size int64) error
	DeleteFile(ctx context.Context, fileName string) error
	ListFiles(ctx context.Context) ([]string, error)
}
