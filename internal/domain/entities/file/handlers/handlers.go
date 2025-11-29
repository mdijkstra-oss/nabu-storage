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
			dispatch.LimitOnAction(file.CreateFile,
				registry.ValidateDomain[file.CreateFilePayload](
					reg,
					validateNoDuplicateCodebook,
					dispatch.ToCreateEntityEvent[file.CreateFilePayload, file.CreatedFilePayload](file.CreateFile, file.CreatedFile, createFileFromPayload),
				),
			),
			dispatch.ToUpdateEntityEvent[file.UpdateFilePayload, file.UpdatedFilePayload](file.UpdateFile, file.UpdatedFile),

			dispatch.LimitOnAction(file.ReplaceFileContent,
				registry.ValidateDomain[file.ReplaceFileContentPayload](
					reg,
					validateFileNotLocked,
					dispatch.ToUpdateEntityEvent[file.ReplaceFileContentPayload, file.ReplacedFileContentPayload](
						file.ReplaceFileContent,
						file.ReplacedFileContent,
						toReplacedContentPayload,
					),
				),
			),

			dispatch.LimitOnAction(file.EditFileContent,
				registry.TransformDomain(
					reg,
					transformEditFileContent,
					dispatch.ToUpdateEntityEvent[file.ReplacedFileContentPayload, file.ReplacedFileContentPayload](
						file.EditFileContent,
						file.ReplacedFileContent,
					),
				),
			),

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
					dispatch.ToUpdateEntityEvent[file.UpdateCodeSectionsPayload, file.UpdatedCodeSectionsPayload](
						file.UpdateCodeSections,
						file.UpdatedCodeSections,
						toUpdatedSectionsPayload,
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
	fileType := payload.Type
	if fileType == "" {
		fileType = file.FileTypeCorpus
	}

	return file.CreatedFilePayload{
		FileData: file.FileData{
			ProjectID:   payload.ProjectID,
			Name:        payload.Name,
			Description: payload.Description,
			Type:        fileType,
			Locked:      fileType.IsLocked(),
		},
		Chunks: createChunksForType(fileType, payload.Content),
	}
}

func createChunksForType(fileType file.FileType, content string) []file.Chunk {
	if fileType.IsChunked() {
		return chunkContent(content)
	}
	return singleChunk(content)
}

func chunkContent(content string) []file.Chunk {
	blocks := chunker.ChunkBlocks(content, chunker.FullPage*5, (chunker.FullPage*5)+chunker.HalfPage)
	return utils.MapWithIndex(blocks, func(i int, block string) file.Chunk {
		return file.Chunk{
			ID:      fmt.Sprintf("%d", i+1),
			Content: block,
			Codes:   []file.CodedSection{},
		}
	})
}

func singleChunk(content string) []file.Chunk {
	return []file.Chunk{{
		ID:      "1",
		Content: content,
		Codes:   []file.CodedSection{},
	}}
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

	if len(failures) == len(payload.Sections) {
		return file.AddCodeSectionsPayload{}, utils.ArrayItemErrors("sections", failures)
	}

	var resultFailures map[int]string
	if len(failures) > 0 {
		resultFailures = failures
	}

	return file.AddCodeSectionsPayload{
		ChunkID:  payload.ChunkID,
		Sections: normalizedSections,
		Failures: resultFailures,
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
		Failures: payload.Failures,
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

	if len(failures) == len(payload.Sections) {
		return file.UpdateCodeSectionsPayload{}, utils.ArrayItemErrors("sections", failures)
	}

	var resultFailures map[int]string
	if len(failures) > 0 {
		resultFailures = failures
	}

	return file.UpdateCodeSectionsPayload{
		ChunkID:  payload.ChunkID,
		Sections: normalizedSections,
		Failures: resultFailures,
	}, nil
}

func toUpdatedSectionsPayload(payload *file.UpdateCodeSectionsPayload) file.UpdatedCodeSectionsPayload {
	return file.UpdatedCodeSectionsPayload{
		ChunkID:  payload.ChunkID,
		Sections: payload.Sections,
		Failures: payload.Failures,
	}
}

func validateNoDuplicateCodebook(proj project.Project, payload file.CreateFilePayload, msg *commands.AnyMessage) error {
	if payload.Type != file.FileTypeCodebook {
		return nil
	}
	if fileview.GetCodebook(proj) != nil {
		return utils.FieldError("type", "project already has a codebook")
	}
	return nil
}

func validateFileNotLocked(proj project.Project, payload file.ReplaceFileContentPayload, msg *commands.AnyMessage) error {
	return errIfLocked(proj, msg)
}

func errIfLocked(proj project.Project, msg *commands.AnyMessage) error {
	if proj.Files[msg.AggregateID].Locked {
		return utils.FieldError("file", "file is locked")
	}
	return nil
}

func toReplacedContentPayload(payload *file.ReplaceFileContentPayload) file.ReplacedFileContentPayload {
	return file.ReplacedFileContentPayload{Content: payload.Content}
}

func transformEditFileContent(proj project.Project, payload file.EditFileContentPayload, msg *commands.AnyMessage) (file.ReplacedFileContentPayload, error) {
	if err := errIfLocked(proj, msg); err != nil {
		return file.ReplacedFileContentPayload{}, err
	}

	f := proj.Files[msg.AggregateID]
	if len(f.Chunks) == 0 {
		return file.ReplacedFileContentPayload{}, utils.FieldError("file", "has no content")
	}

	foundText, found := find.Find(payload.OldText, f.Chunks[0].Content)
	if !found {
		return file.ReplacedFileContentPayload{}, utils.FieldError("old_text", "not found in file")
	}

	return file.ReplacedFileContentPayload{
		Content: find.Replace(f.Chunks[0].Content, foundText, payload.NewText),
	}, nil
}
