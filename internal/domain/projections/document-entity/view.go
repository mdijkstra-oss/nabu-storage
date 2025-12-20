package documentview

import "hermes-relay/internal/domain/entities/document"

type Document = document.Document

type DocumentSummary struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Healthy bool   `json:"healthy"`
	document.DocumentData
}

func ToSummary(d Document) DocumentSummary {
	return DocumentSummary{
		ID:           d.ID,
		Version:      d.Version,
		Healthy:      d.Healthy,
		DocumentData: d.DocumentData,
	}
}
