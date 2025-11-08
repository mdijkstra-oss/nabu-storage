package chunk

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/file"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	"hermes-relay/internal/lib/utils"
	"strconv"
)

type ChunkQuery struct {
	projection.GetByIDQuery
	ChunkFilter
	ID *int `query:"chunkId"`
}

type ChunkResult struct {
	Chunk       file.Chunk `json:"chunk"`
	ChunkID     int        `json:"chunk_id"`
	TotalChunks int        `json:"total_chunks"`
	Next        *int       `json:"next"`
}

func ByChunk(files []fileview.File, q ChunkQuery) []ChunkResult {
	filtered := projection.ByID(files, q.GetByIDQuery)
	if len(filtered) == 0 {
		return nil
	}

	fileData := filtered[0]
	chunks := ApplyFilter(fileData.Chunks, q.ChunkFilter)

	if len(chunks) == 0 {
		return nil
	}

	chunkID := findChunkIndex(chunks, q.ID)
	if chunkID == -1 {
		return nil
	}

	currentChunk := chunks[chunkID]

	var next *int
	if chunkID+1 < len(chunks) {
		nextID := parseID(chunks[chunkID+1].ID)
		next = &nextID
	}

	return []ChunkResult{{
		Chunk:       currentChunk,
		ChunkID:     parseID(currentChunk.ID),
		TotalChunks: len(chunks),
		Next:        next,
	}}
}

func findChunkIndex(chunks []file.Chunk, requestedID *int) int {
	if requestedID == nil {
		return 0
	}

	return utils.FindIndex(chunks, func(c file.Chunk) bool {
		return parseID(c.ID) == *requestedID
	})
}

func parseID(id string) int {
	n, _ := strconv.Atoi(id)
	return n
}
