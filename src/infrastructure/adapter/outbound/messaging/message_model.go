package messaging

import "time"

// Message representa un mensaje básico
type Message struct {
	Body        []byte
	ContentType string
	RoutingKey  string
}

// Work representa un trabajo con headers
type Work struct {
	Message
	Headers map[string]interface{}
}

// MediaTransformationEvent representa un evento de transformación multimedia
type MediaTransformationEvent struct {
	// ========== IDENTIFICACIÓN Y TRAZABILIDAD ==========
	EventID       string    `json:"event_id"`       // UUID único del evento
	RequestID     string    `json:"request_id"`     // UUID de la petición original
	CorrelationID string    `json:"correlation_id"` // UUID para agrupar eventos relacionados
	Timestamp     time.Time `json:"timestamp"`      // Momento del evento
	Version       string    `json:"version"`        // Versión del schema (ej: "1.0")

	// ========== ORIGEN DEL TRABAJO ==========
	SourceService string `json:"source_service"` // "ms-storage-orchestrator"
	UserUUID      string `json:"user_uuid"`      // Quién solicitó
	CompanyID     string `json:"company_id"`     // Empresa
	ResourceID    string `json:"resource_id"`    // ID del recurso asociado
	ResourceType  string `json:"resource_type"`  // Tipo de recurso

	// ========== DATOS DEL ARCHIVO ORIGEN ==========
	SourceFile SourceFileInfo `json:"source_file"`

	// ========== TRANSFORMACIÓN REQUERIDA ==========
	TransformationType   string                 `json:"transformation_type"`   // "image_resize", "video_transcode", "thumbnail"
	TransformationParams map[string]interface{} `json:"transformation_params"` // Parámetros específicos

	// ========== DESTINO ==========
	TargetBucket string `json:"target_bucket"` // Bucket destino
	TargetPath   string `json:"target_path"`   // Path destino

	// ========== CALLBACKS ==========
	CallbackURL string `json:"callback_url,omitempty"` // URL para notificar resultado
	RetryCount  int    `json:"retry_count"`            // Número de reintentos
	MaxRetries  int    `json:"max_retries"`            // Máximo de reintentos
	Priority    int    `json:"priority"`               // Prioridad (1-10)
}

// SourceFileInfo contiene información del archivo origen
type SourceFileInfo struct {
	Bucket      string `json:"bucket"`
	Path        string `json:"path"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
	FileHash    string `json:"file_hash"` // SHA256
}

// TransformationResult representa el resultado de una transformación
type TransformationResult struct {
	// ========== TRAZABILIDAD ==========
	EventID       string    `json:"event_id"`       // Mismo del evento original
	RequestID     string    `json:"request_id"`     // Mismo del evento original
	CorrelationID string    `json:"correlation_id"` // Mismo del evento original
	Timestamp     time.Time `json:"timestamp"`      // Momento de completar

	// ========== ESTADO ==========
	Status       string `json:"status"` // "success", "failed", "partial"
	ErrorMessage string `json:"error_message,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`

	// ========== ARCHIVO RESULTANTE ==========
	OutputFile *OutputFileInfo `json:"output_file,omitempty"`

	// ========== MÉTRICAS ==========
	ProcessingTimeMs int64  `json:"processing_time_ms"` // Tiempo de procesamiento
	WorkerID         string `json:"worker_id"`          // ID del worker que procesó
}

// OutputFileInfo contiene información del archivo generado
type OutputFileInfo struct {
	Bucket      string `json:"bucket"`
	Path        string `json:"path"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	ContentType string `json:"content_type"`
	FileHash    string `json:"file_hash"`
}

// Ejemplos de TransformationParams por tipo:

// Para image_resize:
// {
//   "width": 800,
//   "height": 600,
//   "format": "webp",
//   "quality": 85,
//   "maintain_aspect_ratio": true
// }

// Para video_transcode:
// {
//   "codec": "h264",
//   "resolution": "1080p",
//   "bitrate": "2000k",
//   "format": "mp4",
//   "audio_codec": "aac"
// }

// Para thumbnail:
// {
//   "width": 200,
//   "height": 200,
//   "format": "jpeg",
//   "timestamp_seconds": 5  // para video
// }
