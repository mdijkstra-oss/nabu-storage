package templates

import (
	_ "embed"
)

//go:embed codebook.md
var codebookContent string

//go:embed llm-memo.md
var memoContent string

type DefaultDocument struct {
	Name    string
	Content string
}

func DefaultDocuments() []DefaultDocument {
	return []DefaultDocument{
		{
			Name:    "Codebook",
			Content: codebookContent,
		},
		{
			Name:    "AI Analytical Memo",
			Content: memoContent,
		},
	}
}
