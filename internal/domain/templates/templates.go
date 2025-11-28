package templates

import (
	_ "embed"
	"hermes-relay/internal/domain/entities/file"
)

//go:embed codebook.md
var codebookContent string

type DefaultFile struct {
	Name    string
	Type    file.FileType
	Content string
}

func DefaultFiles() []DefaultFile {
	return []DefaultFile{
		{
			Name:    "Codebook",
			Type:    file.FileTypeCodebook,
			Content: codebookContent,
		},
	}
}
