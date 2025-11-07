package chunk

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/file"
	fileview "hermes-relay/internal/domain/projections/file-entity"
)

type ChunkQuery struct {
	projection.GetByIDQuery
	Index int `path:"index" validate:"gte=0"`
}

type ChunkResult struct {
	Chunk       file.Chunk `json:"chunk"`
	ChunkIndex  int        `json:"chunk_index"`
	TotalChunks int        `json:"total_chunks"`
	HasNext     bool       `json:"has_next"`
}

func ByChunk(files []fileview.File, q ChunkQuery) []ChunkResult {
	filtered := projection.ByID(files, q.GetByIDQuery)

	if len(filtered) == 0 {
		return nil
	}

	file := filtered[0]
	idx := q.Index - 1

	if idx >= len(file.Chunks) || idx < 0 {
		return nil
	}

	return []ChunkResult{{
		Chunk:       file.Chunks[idx],
		ChunkIndex:  q.Index,
		TotalChunks: len(file.Chunks),
		HasNext:     idx < len(file.Chunks)-1,
	}}
}
