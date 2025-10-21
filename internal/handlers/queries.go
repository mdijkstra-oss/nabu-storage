package handlers

import (
	"encoding/json"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/projection"
	"hermes-relay/internal/utils"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func RESTHandler[T any](store *projection.ProjectionStore[T]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, r.Pattern)

		slog.Debug(
			"REST request",
			"path", path,
			"pattern", r.Pattern)

		if path == "" || path == "/" {
			// GET /things
			items, err := projection.GetAll[T](store)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			utils.WarnErr(json.NewEncoder(w).Encode(items))
			return
		}

		// GET /things/{id}
		id := strings.TrimPrefix(path, "/")

		// Check if this is a file query with chunk_index parameter
		if isFileStore(store) && r.URL.Query().Has("chunk_index") {
			handleFileChunk(w, r, id, store)
			return
		}

		// Standard single item query
		item, err := projection.GetByID[T](store, id)
		if err != nil {
			http.Error(w, "not found", 404)
			return
		}

		utils.WarnErr(json.NewEncoder(w).Encode(item))
	}
}

// isFileStore checks if the store is a file.File store
func isFileStore[T any](store *projection.ProjectionStore[T]) bool {
	var zero T
	_, ok := any(zero).(file.File)
	return ok
}

// handleFileChunk handles chunked file responses
func handleFileChunk[T any](w http.ResponseWriter, r *http.Request, id string, store *projection.ProjectionStore[T]) {
	// Get the file
	item, err := projection.GetByID[T](store, id)
	if err != nil {
		http.Error(w, "not found", 404)
		return
	}

	// Cast to file.File
	fileItem, ok := any(item).(*file.File)
	if !ok {
		http.Error(w, "invalid file type", 500)
		return
	}

	// Parse chunk_index parameter
	chunkIndexStr := r.URL.Query().Get("chunk_index")
	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil || chunkIndex < 0 {
		http.Error(w, "invalid chunk_index parameter", 400)
		return
	}

	// Split content into markdown blocks
	blocks := utils.CombineBlocks(utils.SplitMarkdownBlocks(fileItem.Content), 6)

	// Check if chunk index is valid
	if chunkIndex >= len(blocks) {
		http.Error(w, "chunk_index out of range", 404)
		return
	}

	// Build response
	chunk := file.FileChunk{
		ID:         fileItem.ID,
		Chunk:      blocks[chunkIndex],
		ChunkIndex: chunkIndex,
	}

	// Set next chunk index if there are more chunks
	if chunkIndex+1 < len(blocks) {
		chunk.NextChunkIndex = chunkIndex + 1
	} else {
		chunk.NextChunkIndex = -1 // Indicates no more chunks
	}

	w.Header().Set("Content-Type", "application/json")
	utils.WarnErr(json.NewEncoder(w).Encode(chunk))
}
