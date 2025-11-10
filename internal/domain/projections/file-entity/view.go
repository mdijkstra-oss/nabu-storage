package fileview

import "hermes-relay/internal/domain/entities/file"

type File = file.File

// FileSummary contains file metadata without chunks
type FileSummary struct {
	file.BaseFile
}
