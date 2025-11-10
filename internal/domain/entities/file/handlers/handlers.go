package handlers

import (
	"fmt"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/registry"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/lib/text-search/chunker"
	"hermes-relay/internal/lib/utils"
)

func NewRouter(_ *registry.ProjectViewRegistry) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		dispatch.LimitOnEntity(file.EntityName,
			dispatch.ToCreateEntityEvent[file.CreateFilePayload, file.CreatedFilePayload](file.CreateFile, file.CreatedFile, func(payload *file.CreateFilePayload) file.CreatedFilePayload {
				blocks := chunker.ChunkBlocks(payload.Content, chunker.FullPage, chunker.FullPage+chunker.HalfPage)

				chunks := utils.MapWithIndex(blocks, func(i int, block string) file.Chunk {
					return file.Chunk{
						ID:      fmt.Sprintf("%d", i+1),
						Content: block,
						Codes:   []file.CodedSection{},
					}
				})

				return file.CreatedFilePayload{
					CreateFilePayload: file.CreateFilePayload{
						ProjectID:   payload.ProjectID,
						Name:        payload.Name,
						Description: payload.Description,
						Content:     payload.Content,
					},
					Type:   file.FileTypeSource,
					Locked: true,
					Chunks: chunks,
				}
			}),
			dispatch.ToUpdateEntityEvent[file.UpdateFilePayload, file.UpdatedFilePayload](file.UpdateFile, file.UpdatedFile),
			dispatch.ToUpdateEntityEvent[file.CodeFilePayload, file.CodedFilePayload](file.CodeFile, file.CodedFile),
			dispatch.ToEmptyDomainEvent(file.ClearCoding, file.ClearedCoding),
		),
		dispatch.ToEmptyDomainEvent(file.DeleteFile, file.DeletedFile),
	)
}
