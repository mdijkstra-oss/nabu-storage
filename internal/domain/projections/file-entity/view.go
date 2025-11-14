package fileview

import "hermes-relay/internal/domain/entities/file"

type File = file.File

// FileSummary contains file metadata without chunks
type FileSummary struct {
	file.BaseFile
}

// ToSummary strips chunks from File for context-efficient responses
func ToSummary(f File) FileSummary {
	return FileSummary{BaseFile: f.BaseFile}
}
