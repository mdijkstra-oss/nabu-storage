package fileview

import "hermes-relay/internal/domain/entities/file"

type File = file.File

// FileSummary contains file metadata without chunks
type FileSummary struct {
	ID      string `json:"id"`
	Healthy bool   `json:"healthy"`
	file.FileData
}

// ToSummary strips chunks from File for context-efficient responses
func ToSummary(f File) FileSummary {
	return FileSummary{
		ID:       f.ID,
		Healthy:  f.Healthy,
		FileData: f.FileData,
	}
}
