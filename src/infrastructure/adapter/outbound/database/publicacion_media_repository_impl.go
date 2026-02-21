package database

import (
	"context"

	ports "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
)

var _ ports.IPublicacionMediaRepository = (*PublicacionMediaRepositoryImpl)(nil)

type PublicacionMediaRepositoryImpl struct {
	postgresClient *PostgresClient
}

func NewPublicacionMediaRepositoryImpl(postgresClient *PostgresClient) *PublicacionMediaRepositoryImpl {
	return &PublicacionMediaRepositoryImpl{
		postgresClient: postgresClient,
	}
}

func (r *PublicacionMediaRepositoryImpl) UpdateMedia(ctx context.Context, publicacionUUID string, mediaData string) error {
	query := `UPDATE publicacion_media SET media_data = $1 WHERE publicacion_uuid = $2`
	_, err := r.postgresClient.Pool.Query(ctx, query, mediaData, publicacionUUID)
	return err
}

func (r *PublicacionMediaRepositoryImpl) GetMediaByPublicacionID(ctx context.Context, publicacionUUID string) ([]string, error) {
	query := `SELECT media_data FROM publicacion_media WHERE publicacion_uuid = $1`
	rows, err := r.postgresClient.Pool.Query(ctx, query, publicacionUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaList []string
	for rows.Next() {
		var mediaData string
		if err := rows.Scan(&mediaData); err != nil {
			return nil, err
		}
		mediaList = append(mediaList, mediaData)
	}

	return mediaList, nil
}
