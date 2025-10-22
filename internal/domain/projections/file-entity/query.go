package fileview

import (
	"context"
	"errors"
	"hermes-relay/internal/lib/markdown"
	"hermes-relay/internal/projection"
)

type ChunkQuery struct {
	ID    string `path:"id" validate:"required"`
	Index int    `path:"index" validate:"required,gte=1"`
}

type ChunkResult struct {
	Chunk       string `json:"chunk"`
	ChunkIndex  int    `json:"chunk_index"`
	TotalChunks int    `json:"total_chunks"`
	HasNext     bool   `json:"has_next"`
}

func GetFileChunk(ctx context.Context, store *projection.Store[File], q ChunkQuery) (*ChunkResult, error) {
	file, err := store.GetByID(q.ID)

	if err != nil {
		return nil, err
	}

	chunks := markdown.ParseBlocks(file.Content, markdown.HalfPage)

	idx := q.Index - 1

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
