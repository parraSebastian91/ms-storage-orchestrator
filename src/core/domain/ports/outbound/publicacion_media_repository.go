package ports

import "context"

type IPublicacionMediaRepository interface {
	UpdateMedia(ctx context.Context, publicacionUUID string, mediaData string) error
	GetMediaByPublicacionID(ctx context.Context, publicacionUUID string) ([]string, error)
}
