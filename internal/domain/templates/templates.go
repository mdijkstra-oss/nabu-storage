package templates

import (
	_ "embed"
)

//go:embed codebook.md
var codebookContent string //nolint:unused

//go:embed llm-memo.md
var memoContent string //nolint:unused

type DefaultDocument struct {
	Name    string
	Content string
}

func DefaultDocuments() []DefaultDocument {
	return []DefaultDocument{
		//{
		//	Name:    "Codebook",
		//	Content: codebookContent,
		//},
		//{
		//	Name:    "AI Analytical Memo",
		//	Content: memoContent,
		//},
	}
}
