package fileview

import (
	"errors"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/projection"
)

type ChunkQuery struct {
	ID    string `path:"id" validate:"required"`
	Index int    `path:"index" validate:"required,gte=1"`
}

type ChunkResult struct {
	Chunk       file.Chunk `json:"chunk"`
	ChunkIndex  int        `json:"chunk_index"`
	TotalChunks int        `json:"total_chunks"`
	HasNext     bool       `json:"has_next"`
}

func GetFileChunk(store *projection.Store[File], q ChunkQuery) (*ChunkResult, error) {
	file, err := store.GetByID(q.ID)

	if err != nil {
		return nil, err
	}

	idx := q.Index - 1

	chunks := file.Chunks

	if idx >= len(chunks) {
		return nil, errors.New("chunk not found")
	}

	return &ChunkResult{
		Chunk:       chunks[idx],
		ChunkIndex:  q.Index, // display index
		TotalChunks: len(chunks),
		HasNext:     idx < len(chunks)-1,
	}, nil
}
