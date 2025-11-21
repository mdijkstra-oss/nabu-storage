package fileview

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
)

func GetByID(files []file.File, id string) *file.File {
	return projection.GetByID(files, id)
}

func Exists(files []file.File, id string) bool {
	return projection.EntityExists(files, id)
}

func FindChunk(chunks []file.Chunk, chunkID string) *file.Chunk {
	for i := range chunks {
		if chunks[i].ID == chunkID {
			return &chunks[i]
		}
	}
	return nil
}

func GetFileChunk(proj project.Project, fileID, chunkID string) (*file.Chunk, error) {
	fileEntity, exists := proj.Files[fileID]
	if !exists {
		return nil, utils.FieldNotFound("file_id")
	}

	chunk := FindChunk(fileEntity.Chunks, chunkID)
	if chunk == nil {
		return nil, utils.FieldNotFound("chunk_id")
	}

	return chunk, nil
}
