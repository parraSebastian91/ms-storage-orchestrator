// =============================================================================
// media_access_service.go - Servicio de validación de permisos de media
// Integración con la cadena de permisos en PostgreSQL
// =============================================================================

package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
)

// IMediaAccessService define la interfaz de verificación de permisos
type IMediaAccessService interface {
	CheckMediaAccess(ctx context.Context, mediaID uuid.UUID, usuarioUUID uuid.UUID, permiso string) (bool, error)
	LogMediaAccess(ctx context.Context, req LogMediaAccessRequest) (uuid.UUID, error)
	GrantMediaAccess(ctx context.Context, req GrantMediaAccessRequest) (uuid.UUID, error)
	RevokeMediaAccess(ctx context.Context, mediaID uuid.UUID, granteeUUID uuid.UUID, permiso string) error
}

// LogMediaAccessRequest representa una entrada de auditoría
type LogMediaAccessRequest struct {
	MediaID        uuid.UUID
	UsuarioUUID    uuid.UUID
	OrganizacionID int64
	Accion         string // VIEW, DOWNLOAD, SHARE, DELETE
	IPAddress      string
	UserAgent      string
	CorrelationID  string
}

// GrantMediaAccessRequest para otorgar permisos
type GrantMediaAccessRequest struct {
	MediaID        uuid.UUID
	GranteeUUID    uuid.UUID // usuario receptor (opcional)
	GranteeGroupID uuid.UUID // grupo receptor (opcional)
	GranterUUID    uuid.UUID // usuario que otorga (propietario)
	OrganizacionID int64
	Permiso        string
	RazonAcceso    string
	ExpiresAt      *int64 // Timestamp en milisegundos
}

// MediaAccessService implementa IMediaAccessService
type MediaAccessService struct {
	db     *sql.DB
	logger outbound.ILoggerService
}

// NewMediaAccessService crea una nueva instancia
func NewMediaAccessService(db *sql.DB, logger outbound.ILoggerService) *MediaAccessService {
	return &MediaAccessService{
		db:     db,
		logger: logger,
	}
}

// CheckMediaAccess verifica si un usuario tiene permiso para un media_asset
// Utiliza la función PostgreSQL core.check_media_access() que maneja:
// - Acceso de propietario
// - Permisos individuales
// - Permisos de grupo
// - Compartir cross-org
func (s *MediaAccessService) CheckMediaAccess(
	ctx context.Context,
	mediaID uuid.UUID,
	usuarioUUID uuid.UUID,
	permiso string,
) (bool, error) {
	var hasAccess bool

	query := `
		SELECT core.check_media_access($1::UUID, $2::UUID, $3::VARCHAR)
	`

	err := s.db.QueryRowContext(ctx, query, mediaID, usuarioUUID, permiso).Scan(&hasAccess)
	if err != nil {
		s.logger.Error("Error verificando acceso a media", map[string]interface{}{
			"mediaID":     mediaID,
			"usuarioUUID": usuarioUUID,
			"permiso":     permiso,
			"error":       err.Error(),
		})
		return false, err
	}

	return hasAccess, nil
}

// LogMediaAccess registra un acceso en la tabla de auditoría
func (s *MediaAccessService) LogMediaAccess(
	ctx context.Context,
	req LogMediaAccessRequest,
) (uuid.UUID, error) {
	var auditID uuid.UUID

	query := `
		SELECT core.log_media_access(
			$1::UUID,         -- media_id
			$2::UUID,         -- usuario_uuid
			$3::BIGINT,       -- organizacion_id
			$4::VARCHAR,      -- accion
			$5::VARCHAR,      -- ip_address
			$6::TEXT,         -- user_agent
			$7::VARCHAR       -- correlation_id
		)
	`

	err := s.db.QueryRowContext(
		ctx,
		query,
		req.MediaID,
		req.UsuarioUUID,
		req.OrganizacionID,
		req.Accion,
		req.IPAddress,
		req.UserAgent,
		req.CorrelationID,
	).Scan(&auditID)

	if err != nil {
		s.logger.Error("Error registrando acceso a media", map[string]interface{}{
			"mediaID":       req.MediaID,
			"usuarioUUID":   req.UsuarioUUID,
			"accion":        req.Accion,
			"error":         err.Error(),
			"correlationId": req.CorrelationID,
		})
		return uuid.Nil, err
	}

	s.logger.Info("Acceso a media registrado", map[string]interface{}{
		"auditID":       auditID,
		"mediaID":       req.MediaID,
		"usuarioUUID":   req.UsuarioUUID,
		"accion":        req.Accion,
		"correlationId": req.CorrelationID,
	})

	return auditID, nil
}

// GrantMediaAccess otorga permisos a un usuario o grupo para acceder a un media_asset
func (s *MediaAccessService) GrantMediaAccess(
	ctx context.Context,
	req GrantMediaAccessRequest,
) (uuid.UUID, error) {
	var policyID uuid.UUID

	query := `
		INSERT INTO core.media_access_policy (
			media_id,
			grantee_usuario_uuid,
			grantee_grupo_id,
			granter_usuario_uuid,
			organizacion_id,
			permiso,
			razon_acceso,
			expires_at
		) VALUES (
			$1::UUID,      -- media_id
			$2::UUID,      -- grantee_usuario_uuid (nullable)
			$3::UUID,      -- grantee_grupo_id (nullable)
			$4::UUID,      -- granter_usuario_uuid
			$5::BIGINT,    -- organizacion_id
			$6::VARCHAR,   -- permiso
			$7::VARCHAR,   -- razon_acceso
			to_timestamp($8::BIGINT / 1000.0)::TIMESTAMPTZ  -- expires_at
		)
		ON CONFLICT (media_id, grantee_usuario_uuid, permiso) 
			DO UPDATE SET revoked_at = NULL
		RETURNING id
	`

	var expiresAt interface{} = nil
	if req.ExpiresAt != nil {
		expiresAt = *req.ExpiresAt
	}

	err := s.db.QueryRowContext(
		ctx,
		query,
		req.MediaID,
		nullUUID(req.GranteeUUID),
		nullUUID(req.GranteeGroupID),
		req.GranterUUID,
		req.OrganizacionID,
		req.Permiso,
		req.RazonAcceso,
		expiresAt,
	).Scan(&policyID)

	if err != nil {
		s.logger.Error("Error otorgando permiso de media", map[string]interface{}{
			"mediaID":        req.MediaID,
			"granteeUUID":    req.GranteeUUID,
			"granteeGroupID": req.GranteeGroupID,
			"permiso":        req.Permiso,
			"error":          err.Error(),
		})
		return uuid.Nil, err
	}

	s.logger.Info("Permiso de media otorgado", map[string]interface{}{
		"policyID":       policyID,
		"mediaID":        req.MediaID,
		"granteeUUID":    req.GranteeUUID,
		"granteeGroupID": req.GranteeGroupID,
		"permiso":        req.Permiso,
	})

	return policyID, nil
}

// RevokeMediaAccess revoca el acceso a un media_asset
func (s *MediaAccessService) RevokeMediaAccess(
	ctx context.Context,
	mediaID uuid.UUID,
	granteeUUID uuid.UUID,
	permiso string,
) error {
	query := `
		UPDATE core.media_access_policy
		SET revoked_at = NOW()
		WHERE media_id = $1::UUID
		  AND grantee_usuario_uuid = $2::UUID
		  AND permiso = $3::VARCHAR
		  AND revoked_at IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, mediaID, granteeUUID, permiso)
	if err != nil {
		s.logger.Error("Error revocando permiso de media", map[string]interface{}{
			"mediaID":     mediaID,
			"granteeUUID": granteeUUID,
			"permiso":     permiso,
			"error":       err.Error(),
		})
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	s.logger.Info("Permiso de media revocado", map[string]interface{}{
		"mediaID":      mediaID,
		"granteeUUID":  granteeUUID,
		"permiso":      permiso,
		"rowsAffected": rowsAffected,
	})

	return nil
}

// Helper: convertir UUID vacío a NULL para PostgreSQL
func nullUUID(u uuid.UUID) interface{} {
	if u == uuid.Nil {
		return nil
	}
	return u
}

// =============================================================================
// INTEGRACIÓN EN STORAGE_CONTROLLER.GO
// =============================================================================

// Ejemplo: Modificar GetPresignedGetURL para validar permisos

/*
func (c *StorageController) GetPresignedGetURL(ctx fiber.Ctx) error {
	start := time.Now()
	var presignedURLRequest inbound_dto.PresignedGetURLRequestDTO
	if err := ctx.Bind().Query(&presignedURLRequest); err != nil {
		c.logger.Error("Error al parsear la solicitud", map[string]interface{}{
			"error": err.Error(),
		})
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Faltan Datos",
		})
	}

	correlationId := strings.TrimSpace(ctx.Get("X-Correlation-Id"))
	if correlationId == "" {
		correlationId = "N/A"
	}

	// NUEVO: Extraer usuario de JWT (del contexto de auth)
	usuarioUUID, err := c.extractUsuarioUUID(ctx)
	if err != nil {
		c.logger.Warn("No se pudo extraer usuario", map[string]interface{}{
			"correlationId": correlationId,
			"error":         err.Error(),
		})
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Usuario no identificado",
		})
	}

	// NUEVO: Obtener media_id desde storage_key (necesita parsear la ruta)
	mediaID := c.extractMediaIDFromStorageKey(presignedURLRequest.Storage_key)
	organizacionID := c.extractOrgFromContext(ctx) // O del JWT

	// NUEVO: Verificar permiso antes de generar URL
	hasAccess, err := c.mediaAccessService.CheckMediaAccess(
		ctx.Context(),
		mediaID,
		usuarioUUID,
		"VIEW",  // o "DOWNLOAD" según la acción
	)

	if err != nil {
		c.logger.Error("Error verificando acceso", map[string]interface{}{
			"mediaID":       mediaID,
			"usuarioUUID":   usuarioUUID,
			"correlationId": correlationId,
			"error":         err.Error(),
		})
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error interno",
		})
	}

	if !hasAccess {
		c.logger.Warn("Acceso denegado a media", map[string]interface{}{
			"mediaID":       mediaID,
			"usuarioUUID":   usuarioUUID,
			"storage_key":   presignedURLRequest.Storage_key,
			"correlationId": correlationId,
		})

		// NUEVO: Registrar intento de acceso no autorizado
		c.mediaAccessService.LogMediaAccess(ctx.Context(), service.LogMediaAccessRequest{
			MediaID:        mediaID,
			UsuarioUUID:    usuarioUUID,
			OrganizacionID: organizacionID,
			Accion:         "UNAUTHORIZED_ATTEMPT",
			IPAddress:      ctx.IP(),
			UserAgent:      ctx.Get("User-Agent"),
			CorrelationID:  correlationId,
		})

		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "No tienes permisos para acceder a este recurso",
		})
	}

	// Generar URL presignada
	url, err := c.storageApplication.ExecuteGetPresignedGetURL(
		ctx.Context(),
		command.GetPresignedGetURLCommand{
			Storage_key:   presignedURLRequest.Storage_key,
			CorrelationId: correlationId,
		},
	)

	if err != nil {
		c.logger.Error("Error al obtener URL", map[string]interface{}{
			"error":         err.Error(),
			"correlationId": correlationId,
		})
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error",
		})
	}

	// NUEVO: Registrar acceso exitoso
	c.mediaAccessService.LogMediaAccess(ctx.Context(), service.LogMediaAccessRequest{
		MediaID:        mediaID,
		UsuarioUUID:    usuarioUUID,
		OrganizacionID: organizacionID,
		Accion:         "DOWNLOAD",
		IPAddress:      ctx.IP(),
		UserAgent:      ctx.Get("User-Agent"),
		CorrelationID:  correlationId,
	})

	c.logger.Info("URL presignada generada", map[string]interface{}{
		"mediaID":       mediaID,
		"usuarioUUID":   usuarioUUID,
		"correlationId": correlationId,
		"durationMs":    time.Since(start).Milliseconds(),
	})

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"url":           url,
		"correlationId": correlationId,
	})
}
*/
