package templates

import (
	_ "embed"
	"hermes-relay/internal/domain/entities/file"
)

//go:embed codebook.md
var codebookContent string

//go:embed llm-memo.md
var memoContent string

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
		{
			Name:    "AI Analytical Memo",
			Type:    file.FileTypeLLMMemo,
			Content: memoContent,
		},
	}
}
