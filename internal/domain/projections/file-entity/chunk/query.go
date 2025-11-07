package chunk

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/file"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	"strconv"
)

type ChunkQuery struct {
	projection.GetByIDQuery
	ChunkFilter
	IDX *int `query:"index"`
}

type ChunkResult struct {
	Chunk       file.Chunk `json:"chunk"`
	ChunkIndex  int        `json:"chunk_index"`
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

	chunkIdx := findChunkIndex(chunks, q.IDX)
	if chunkIdx == -1 {
		return nil
	}

	currentChunk := chunks[chunkIdx]

	var next *int
	if chunkIdx+1 < len(chunks) {
		nextIdx := parseIDX(chunks[chunkIdx+1].IDX)
		next = &nextIdx
	}

	return []ChunkResult{{
		Chunk:       currentChunk,
		ChunkIndex:  parseIDX(currentChunk.IDX),
		TotalChunks: len(chunks),
		Next:        next,
	}}
}

func findChunkIndex(chunks []file.Chunk, requestedIDX *int) int {
	if requestedIDX == nil {
		return 0
	}

	for i, chunk := range chunks {
		if parseIDX(chunk.IDX) == *requestedIDX {
			return i
		}
	}

	return -1
}

func parseIDX(idx string) int {
	n, _ := strconv.Atoi(idx)
	return n
}
