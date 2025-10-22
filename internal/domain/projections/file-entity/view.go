package fileview

import "hermes-relay/internal/domain/entities/file"

type File = file.File

// Claude: Used in /file/{id}/chunk/{chunkIndex} query
type FileChunk struct {
	ID     string   `json:"id" validate:"required"`
	Chunks []string `json:"chunk"`
	//ChunkIndex     int    `json:"chunk_index" validate:"required"`
	//NextChunkIndex int    `json:"next_chunk_index"`
}
