package command

import "io"

type UploadFileCommand struct {
	FileData     io.Reader
	FileMetadata FileMetadata
}

type FileMetadata struct {
	UUID          string // UUID del usuario/propietario
	FileName      string
	FileSize      int
	ContentType   string
	Tags          string
	Description   string
	RetentionDays int
	IsPublic      bool
}
