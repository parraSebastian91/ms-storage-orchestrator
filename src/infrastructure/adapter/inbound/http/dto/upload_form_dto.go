package dto

type UploadFormDTO struct {
	// ========== IDENTIFICACIÓN DEL USUARIO/PROPIETARIO ==========
	UserUUID     string `form:"user_uuid" validate:"required,uuid4"`
	CompanyID    string `form:"company_id" validate:"required,uuid4"`
	DepartmentID string `form:"department_id" validate:"required,uuid4"`

	// ========== IDENTIFICACIÓN DE LA APLICACIÓN ==========
	ApplicationID string `form:"application_id" validate:"required"` // ej: "ERP", "CMS", "BILLING"
	ModuleID      string `form:"module_id" validate:"required"`      // ej: "invoices", "documents", "media"
	ProcessType   string `form:"process_type" validate:"required"`   // ej: "invoice_upload", "video_transcode"

	// ========== IDENTIFICACIÓN DEL RECURSO/CONTEXTO ==========
	ResourceID   string `form:"resource_id" validate:"required,uuid4"` // ej: invoice_id, order_id
	ResourceType string `form:"resource_type" validate:"required"`     // ej: "invoice", "order", "contract"

	// ========== DATOS DEL ARCHIVO ==========
	FileName    string `form:"file_name" validate:"required"`
	FileSize    int64  `form:"file_size" validate:"required,min=1"`
	ContentType string `form:"content_type" validate:"required"` // ej: "image/jpeg", "application/pdf"
	FileHash    string `form:"file_hash" validate:"omitempty"`   // SHA256 para deduplicación

	// ========== METADATA ADICIONAL ==========
	Tags          string `form:"tags" validate:"omitempty"` // JSON array: ["invoice", "urgent"]
	Description   string `form:"description" validate:"omitempty,max=500"`
	RetentionDays int    `form:"retention_days" validate:"omitempty,min=1"` // Cuántos días guardar
	IsPublic      bool   `form:"is_public" validate:"omitempty"`            // Acceso público o privado
}

// Estructura para almacenar en BD (metadatos del archivo)
type FileMetadata struct {
	ID           string `json:"id" db:"id"` // UUID
	UserUUID     string `json:"user_uuid" db:"user_uuid"`
	CompanyID    string `json:"company_id" db:"company_id"`
	ResourceID   string `json:"resource_id" db:"resource_id"`
	ResourceType string `json:"resource_type" db:"resource_type"`

	ObjectPath  string `json:"object_path" db:"object_path"` // Path en MinIO: /company/user/resource/file.ext
	FileName    string `json:"file_name" db:"file_name"`
	FileSize    int64  `json:"file_size" db:"file_size"`
	ContentType string `json:"content_type" db:"content_type"`
	FileHash    string `json:"file_hash" db:"file_hash"`

	ApplicationID string `json:"application_id" db:"application_id"`
	ModuleID      string `json:"module_id" db:"module_id"`
	ProcessType   string `json:"process_type" db:"process_type"`

	Tags        string `json:"tags" db:"tags"`
	Description string `json:"description" db:"description"`

	Status     string `json:"status" db:"status"` // pending, processing, completed, failed
	UploadedAt string `json:"uploaded_at" db:"uploaded_at"`
	ExpiresAt  string `json:"expires_at" db:"expires_at"`

	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

// storage-bucket-raw/
// ├── company-uuid/
// │   ├── user-uuid/
// │   │   ├── resource-type/
// │   │   │   ├── resource-id/
// │   │   │   │   ├── file-uuid-filename.ext

// -> /12345678-company/87654321-user/invoice/99999-resource/abc123-invoice.pdf
