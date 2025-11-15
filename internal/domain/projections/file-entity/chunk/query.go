package chunk

import (
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/lib/utils"
	"strconv"
)

type ChunkQuery struct {
	ChunkFilter
	ID      string `path:"id" validate:"required,valid_id"`
	ChunkID string `query:"chunkId"`
}

type ChunkResult struct {
	Chunk       file.Chunk `json:"chunk"`
	ChunkID     int        `json:"chunk_id"`
	TotalChunks int        `json:"total_chunks"`
	Next        *int       `json:"next"`
}

// GetChunk applies query to file and returns single chunk with navigation
func GetChunk(f file.File, q ChunkQuery) *ChunkResult {
	chunks := ApplyFilter(f.Chunks, q.ChunkFilter)

	if len(chunks) == 0 {
		return nil
	}

	chunkIndex := findChunkIndex(chunks, q.ChunkID)
	if chunkIndex == -1 {
		return nil
	}

	currentChunk := chunks[chunkIndex]

	var next *int
	if chunkIndex+1 < len(chunks) {
		nextID := parseID(chunks[chunkIndex+1].ID)
		next = &nextID
	}

	return &ChunkResult{
		Chunk:       currentChunk,
		ChunkID:     parseID(currentChunk.ID),
		TotalChunks: len(chunks),
		Next:        next,
	}
}

func findChunkIndex(chunks []file.Chunk, requestedID string) int {
	if requestedID == "" {
		return 0
	}

	return utils.FindIndex(chunks, func(c file.Chunk) bool {
		return c.ID == requestedID
	})
}

func parseID(id string) int {
	n, _ := strconv.Atoi(id)
	return n
}
