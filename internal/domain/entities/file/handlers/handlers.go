package handlers

import (
	"fmt"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/lib/text-search/chunker"
	"hermes-relay/internal/lib/utils"
)

func NewRouter(_ *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		dispatch.LimitOnEntity(file.EntityName,
			dispatch.ToCreateEntityEvent[file.CreateFilePayload, file.CreatedFilePayload](file.CreateFile, file.CreatedFile, createFileFromPayload),
			dispatch.ToUpdateEntityEvent[file.UpdateFilePayload, file.UpdatedFilePayload](file.UpdateFile, file.UpdatedFile),
			dispatch.ToUpdateEntityEvent[file.AddCodeSectionsPayload, file.AddedCodeSectionsPayload](file.AddCodeSections, file.AddedCodeSections, assignSectionIDs),
			dispatch.ToUpdateEntityEvent[file.UpdateCodeSectionsPayload, file.UpdateCodeSectionsPayload](file.UpdateCodeSections, file.UpdatedCodeSections),
			dispatch.ToUpdateEntityEvent[file.RemoveCodeSectionsPayload, file.RemoveCodeSectionsPayload](file.RemoveCodeSections, file.RemovedCodeSections),
			dispatch.ToEmptyDomainEvent(file.ClearCoding, file.ClearedCoding),
		),
		dispatch.ToEmptyDomainEvent(file.DeleteFile, file.DeletedFile),
	)
}

func createFileFromPayload(payload *file.CreateFilePayload) file.CreatedFilePayload {
	return file.CreatedFilePayload{
		FileData: file.FileData{
			ProjectID:   payload.ProjectID,
			Name:        payload.Name,
			Description: payload.Description,
			Type:        file.FileTypeSource,
			Locked:      true,
		},
		Chunks: createFileChunks(payload.Content),
	}
}

func createFileChunks(content string) []file.Chunk {
	// Warn! These will be now fixed in time.
	// Changing it later if ever, will only be for future files. Perhaps not that bad....
	blocks := chunker.ChunkBlocks(content, chunker.FullPage*5, (chunker.FullPage*5)+chunker.HalfPage)
	return utils.MapWithIndex(blocks, func(i int, block string) file.Chunk {
		return file.Chunk{
			ID:      fmt.Sprintf("%d", i+1),
			Content: block,
			Codes:   []file.CodedSection{},
		}
	})
}

func assignSectionIDs(payload *file.AddCodeSectionsPayload) file.AddedCodeSectionsPayload {
	sections := utils.Map(payload.Sections, func(op file.AddSectionOp) file.AddedSection {
		return file.AddedSection{
			ID:       utils.NewID(),
			CodeSlug: op.CodeSlug,
			CodeID:   op.CodeID,
			Text:     op.Text,
			Reason:   op.Reason,
		}
	})
	return file.AddedCodeSectionsPayload{
		ChunkID:  payload.ChunkID,
		Sections: sections,
	}
}
