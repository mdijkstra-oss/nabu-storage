package fileview

import "hermes-relay/internal/domain/entities/file"

type File = file.File

// FileSummary contains file metadata without content
type FileSummary struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Healthy bool   `json:"healthy"`
	file.FileData
}

// ToSummary strips content and codes from File for context-efficient responses
func ToSummary(f File) FileSummary {
	return FileSummary{
		ID:       f.ID,
		Version:  f.Version,
		Healthy:  f.Healthy,
		FileData: f.FileData,
	}
}
