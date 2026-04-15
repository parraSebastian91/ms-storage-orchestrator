package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	AplicationModel "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/application/model"
	domainModels "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/models"
	ports "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
)

var _ ports.IMediaRepository = (*MediaRepositoryAdapter)(nil)

type MediaRepositoryAdapter struct {
	postgresClient *PostgresClient
	logger         ports.ILoggerService
}

func NewMediaRepositoryAdapter(postgresClient *PostgresClient, logger ports.ILoggerService) *MediaRepositoryAdapter {
	return &MediaRepositoryAdapter{
		postgresClient: postgresClient,
		logger:         logger,
	}
}

func (a *MediaRepositoryAdapter) CreateMediaMetadata(ctx context.Context, model AplicationModel.StorageModel) error {
	if a == nil || a.postgresClient == nil || a.postgresClient.Pool == nil {
		return fmt.Errorf("postgres client is not initialized")
	}

	start := time.Now()
	a.logger.Info("CreateMediaMetadata started", map[string]interface{}{
		"ownerId":       model.OwnerUUID,
		"mediaType":     model.MediaType,
		"category":      model.CategoryProcess,
		"storageKey":    model.StorageKey,
		"correlationId": model.CorrelationId,
	})

	query := `
		INSERT INTO media.media_assets (
			owner_id,
			m_type,
			category,
			status,
			original_name,
			mime_type,
			storage_key,
			correlation_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := a.postgresClient.Pool.Exec(
		ctx,
		query,
		model.OwnerUUID,
		model.MediaType,
		model.CategoryProcess,
		domainModels.STATE_PROCESS_PENDING,
		model.NameFile,
		model.FormatFile,
		model.StorageKey,
		model.CorrelationId,
	)
	if err != nil {
		a.logger.Error("Failed to create media metadata", map[string]interface{}{
			"error":         err.Error(),
			"ownerId":       model.OwnerUUID,
			"mediaType":     model.MediaType,
			"category":      model.CategoryProcess,
			"storageKey":    model.StorageKey,
			"correlationId": model.CorrelationId,
			"durationMs":    time.Since(start).Milliseconds(),
		})
		return err
	}

	a.logger.Info("Media metadata created successfully", map[string]interface{}{
		"ownerId":       model.OwnerUUID,
		"mediaType":     model.MediaType,
		"category":      model.CategoryProcess,
		"storageKey":    model.StorageKey,
		"correlationId": model.CorrelationId,
		"durationMs":    time.Since(start).Milliseconds(),
	})

	return nil
}

func (a *MediaRepositoryAdapter) GetMediaMetadata(ctx context.Context, objectKey string) (AplicationModel.StorageModel, error) {
	if a == nil || a.postgresClient == nil || a.postgresClient.Pool == nil {
		return AplicationModel.StorageModel{}, fmt.Errorf("postgres client is not initialized")
	}

	start := time.Now()

	normalizedKey := strings.TrimSpace(objectKey)
	decodedKey, err := url.QueryUnescape(normalizedKey)
	if err == nil {
		normalizedKey = decodedKey
	}

	keysToTry := []string{normalizedKey}
	if objectKey != normalizedKey {
		keysToTry = append(keysToTry, strings.TrimSpace(objectKey))
	}

	query := `
		SELECT 
			id,
			owner_id, 
			m_type, 
			category, 
			original_name, 
			mime_type, 
			storage_key,
			correlation_id 
		FROM 
			media.media_assets
		where 
			storage_key = $1`

	for _, key := range keysToTry {
		row := a.postgresClient.Pool.QueryRow(ctx, query, key)
		var mediaModel AplicationModel.StorageModel
		err := row.Scan(
			&mediaModel.AssetId,
			&mediaModel.OwnerUUID,
			&mediaModel.MediaType,
			&mediaModel.CategoryProcess,
			&mediaModel.NameFile,
			&mediaModel.FormatFile,
			&mediaModel.StorageKey,
			&mediaModel.CorrelationId,
		)
		if err == nil {
			a.logger.Info("GetMediaMetadata found record", map[string]interface{}{
				"storageKey":    key,
				"assetId":       mediaModel.AssetId,
				"correlationId": mediaModel.CorrelationId,
				"durationMs":    time.Since(start).Milliseconds(),
			})
			return mediaModel, nil
		}

		if !errors.Is(err, pgx.ErrNoRows) {
			a.logger.Error("GetMediaMetadata query failed", map[string]interface{}{
				"error":      err.Error(),
				"storageKey": key,
			})
			return AplicationModel.StorageModel{}, err
		}
	}

	a.logger.Warn("GetMediaMetadata no rows", map[string]interface{}{
		"storageKey": normalizedKey,
		"durationMs": time.Since(start).Milliseconds(),
	})

	return AplicationModel.StorageModel{}, pgx.ErrNoRows
}
