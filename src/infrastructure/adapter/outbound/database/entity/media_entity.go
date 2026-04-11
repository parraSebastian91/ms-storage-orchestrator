package entity

type MediaAssetsEntity struct {
	Id           string `db:"id"`
	OwnerId      string `db:"owner_id"`
	MType        string `db:"m_type"`   // indica el tipo de media (image, video, document, etc.) , fotoPerfil, banner, documento
	Category     string `db:"category"` //indica a rust que tipo de procesamiento se le debe hacer al archivo
	Status       string `db:"status"`   //estatus del proceso
	OriginalName string `db:"original_name"`
	MimeType     string `db:"mime_type"`
	StorageKey   string `db:"storage_key"`
	ErrorLog     string `db:"error_log"`
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
}

type MediaVariantEntity struct {
	Id          string `db:"id"`
	AssetId     string `db:"asset_id"`
	VariantName string `db:"variant_name"`
	UrlPath     string `db:"url_path"`
	Metadata    string `db:"metadata"` // JSONB como string
	CreatedAt   string `db:"created_at"`
}
