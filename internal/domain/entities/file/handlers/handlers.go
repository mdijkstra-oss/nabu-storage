package handlers

import (
	"fmt"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/lib/text-search/chunker"
	"hermes-relay/internal/lib/text-search/find"
	"hermes-relay/internal/lib/utils"
)

func NewRouter(reg *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		dispatch.LimitOnEntity(file.EntityName,
			dispatch.ToCreateEntityEvent[file.CreateFilePayload, file.CreatedFilePayload](file.CreateFile, file.CreatedFile, createFileFromPayload),
			dispatch.ToUpdateEntityEvent[file.UpdateFilePayload, file.UpdatedFilePayload](file.UpdateFile, file.UpdatedFile),

			dispatch.LimitOnAction(file.AddCodeSections,
				registry.NormalizeDomain[file.AddCodeSectionsPayload](
					reg,
					normalizeAddSections,
					dispatch.ToUpdateEntityEvent[file.AddCodeSectionsPayload, file.AddedCodeSectionsPayload](
						file.AddCodeSections,
						file.AddedCodeSections,
						addSectionIDs,
					),
				),
			),

			dispatch.LimitOnAction(file.UpdateCodeSections,
				registry.NormalizeDomain[file.UpdateCodeSectionsPayload](
					reg,
					normalizeUpdateSections,
					dispatch.ToUpdateEntityEvent[file.UpdateCodeSectionsPayload, file.UpdateCodeSectionsPayload](
						file.UpdateCodeSections,
						file.UpdatedCodeSections,
					),
				),
			),

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

func validateAndNormalizeText(text, chunkContent string) (string, error) {
	if find.CountWords(text) < 3 {
		return "", fmt.Errorf("minimum 3 words required: %q", text)
	}

	normalizedText, found := find.Find(text, chunkContent)
	if !found {
		return "", fmt.Errorf("text not found in chunk: %q", text)
	}

	return normalizedText, nil
}

func normalizeAddSections(proj project.Project, payload file.AddCodeSectionsPayload, msg *commands.AnyMessage) (file.AddCodeSectionsPayload, error) {
	chunk, err := fileview.GetFileChunk(proj, msg.AggregateID, payload.ChunkID)
	if err != nil {
		return file.AddCodeSectionsPayload{}, err
	}

	normalizedSections := []file.AddSectionOp{}
	failures := make(map[int]string)

	for i, op := range payload.Sections {
		normalizedText, err := validateAndNormalizeText(op.Text, chunk.Content)
		if err != nil {
			failures[i] = err.Error()
			continue
		}

		normalizedSections = append(normalizedSections, file.AddSectionOp{
			CodeSlug: op.CodeSlug,
			CodeID:   op.CodeID,
			Text:     normalizedText,
			Reason:   op.Reason,
		})
	}

	if len(failures) > 0 {
		return file.AddCodeSectionsPayload{}, utils.ArrayItemErrors("sections", failures)
	}

	return file.AddCodeSectionsPayload{
		ChunkID:  payload.ChunkID,
		Sections: normalizedSections,
	}, nil
}

func addSectionIDs(payload *file.AddCodeSectionsPayload) file.AddedCodeSectionsPayload {
	return file.AddedCodeSectionsPayload{
		ChunkID: payload.ChunkID,
		Sections: utils.Map(payload.Sections, func(op file.AddSectionOp) file.AddedSection {
			return file.AddedSection{
				ID:       utils.NewID(),
				CodeSlug: op.CodeSlug,
				CodeID:   op.CodeID,
				Text:     op.Text,
				Reason:   op.Reason,
			}
		}),
	}
}

func normalizeUpdateSections(proj project.Project, payload file.UpdateCodeSectionsPayload, msg *commands.AnyMessage) (file.UpdateCodeSectionsPayload, error) {
	chunk, err := fileview.GetFileChunk(proj, msg.AggregateID, payload.ChunkID)
	if err != nil {
		return file.UpdateCodeSectionsPayload{}, err
	}

	normalizedSections := []file.UpdateSectionOp{}
	failures := make(map[int]string)

	for i, op := range payload.Sections {
		normalizedOp := file.UpdateSectionOp{
			ID:     op.ID,
			Reason: op.Reason,
		}

		if op.Text != "" {
			normalizedText, err := validateAndNormalizeText(op.Text, chunk.Content)
			if err != nil {
				failures[i] = err.Error()
				continue
			}
			normalizedOp.Text = normalizedText
		}

		normalizedSections = append(normalizedSections, normalizedOp)
	}

	if len(failures) > 0 {
		return file.UpdateCodeSectionsPayload{}, utils.ArrayItemErrors("sections", failures)
	}

	return file.UpdateCodeSectionsPayload{
		ChunkID:  payload.ChunkID,
		Sections: normalizedSections,
	}, nil
}
