package handlers

import (
	"fmt"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/lib/text-search/find"
	"hermes-relay/internal/lib/utils"
	"strings"
)

func NewRouter(reg *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		dispatch.LimitOnEntity(file.EntityName,
			dispatch.LimitOnAction(file.CreateFile,
				registry.ValidateDomain[file.CreateFilePayload](
					reg,
					validateNoDuplicateSingleton,
					dispatch.ToCreateEntityEvent[file.CreateFilePayload, file.CreatedFilePayload](file.CreateFile, file.CreatedFile, createFileFromPayload),
				),
			),
			dispatch.ToUpdateEntityEvent[file.UpdateFilePayload, file.UpdatedFilePayload](file.UpdateFile, file.UpdatedFile),
			dispatch.ToEmptyDomainEvent(file.PinFile, file.PinnedFile),
			dispatch.ToEmptyDomainEvent(file.UnpinFile, file.UnpinnedFile),

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

			dispatch.LimitOnAction(file.RemoveCodeSections,
				registry.ValidateDomain(
					reg,
					validateCorpusFile[file.RemoveCodeSectionsPayload],
					dispatch.ToUpdateEntityEvent[file.RemoveCodeSectionsPayload, file.RemoveCodeSectionsPayload](file.RemoveCodeSections, file.RemovedCodeSections),
				),
			),

			dispatch.LimitOnAction(file.ClearCoding,
				registry.ValidateDomain(
					reg,
					validateCorpusFile[commands.EmptyPayload],
					dispatch.ToEmptyDomainEvent(file.ClearCoding, file.ClearedCoding),
				),
			),

			dispatch.LimitOnAction(file.RemoveCodeFromFile,
				registry.ValidateDomain(
					reg,
					validateCorpusFile[file.RemoveCodeFromFilePayload],
					dispatch.ToUpdateEntityEvent[file.RemoveCodeFromFilePayload, file.RemovedCodeFromFilePayload](
						file.RemoveCodeFromFile,
						file.RemovedCodeFromFile,
					),
				),
			),
		),
		dispatch.ToEmptyDomainEvent(file.DeleteFile, file.DeletedFile),
	)
}

func createFileFromPayload(payload *file.CreateFilePayload) file.CreatedFilePayload {
	fileType := payload.Type
	if fileType == "" {
		fileType = file.FileTypeCorpus
	}

	content := payload.Content
	if content != "" && !strings.HasSuffix(content, "\n") {
		content = content + "\n"
	}

	return file.CreatedFilePayload{
		FileData: file.FileData{
			ProjectID:   payload.ProjectID,
			Name:        payload.Name,
			Description: payload.Description,
			Type:        fileType,
			Locked:      fileType.IsLocked(),
		},
		Content: content,
		Codes:   []file.CodedSection{},
	}
}

func validateAndNormalizeText(text, fileContent string) (string, error) {
	wordCount := find.CountWords(text)
	if wordCount < 3 {
		return "", fmt.Errorf("text too short (%d words, need 3+) - expand selection: %q", wordCount, text)
	}

	normalizedText, found := find.Find(text, fileContent)
	if !found {
		return "", fmt.Errorf("text not found in file: %q", text)
	}

	return normalizedText, nil
}

func normalizeAddSections(proj project.Project, payload file.AddCodeSectionsPayload, msg *commands.AnyMessage) (file.AddCodeSectionsPayload, error) {
	if err := errIfNotCorpus(proj, msg); err != nil {
		return file.AddCodeSectionsPayload{}, err
	}

	f := proj.Files[msg.AggregateID]

	normalizedSections := []file.AddSectionOp{}
	failures := make(map[int]string)

	for i, op := range payload.Sections {
		normalizedText, err := validateAndNormalizeText(op.Text, f.Content)
		if err != nil {
			failures[i] = err.Error()
			continue
		}

		normalizedSections = append(normalizedSections, file.AddSectionOp{
			CodeID:     op.CodeID,
			Text:       normalizedText,
			Reason:     op.Reason,
			Confidence: op.Confidence,
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
		Sections: normalizedSections,
		Failures: resultFailures,
	}, nil
}

func addSectionIDs(payload *file.AddCodeSectionsPayload) file.AddedCodeSectionsPayload {
	return file.AddedCodeSectionsPayload{
		Sections: utils.Map(payload.Sections, func(op file.AddSectionOp) file.AddedSection {
			return file.AddedSection{
				ID:         utils.NewID(),
				CodeID:     op.CodeID,
				Text:       op.Text,
				Reason:     op.Reason,
				Confidence: op.Confidence,
			}
		}),
		Failures: payload.Failures,
	}
}

func normalizeUpdateSections(proj project.Project, payload file.UpdateCodeSectionsPayload, msg *commands.AnyMessage) (file.UpdateCodeSectionsPayload, error) {
	if err := errIfNotCorpus(proj, msg); err != nil {
		return file.UpdateCodeSectionsPayload{}, err
	}

	f := proj.Files[msg.AggregateID]

	normalizedSections := []file.UpdateSectionOp{}
	failures := make(map[int]string)

	for i, op := range payload.Sections {
		section := fileview.FindSection(f, op.ID)
		if section == nil {
			failures[i] = fmt.Sprintf("section not found: %s", op.ID)
			continue
		}

		normalizedOp := file.UpdateSectionOp{
			ID:         op.ID,
			Reason:     op.Reason,
			Confidence: op.Confidence,
		}

		if op.Text != "" {
			normalizedText, err := validateAndNormalizeText(op.Text, f.Content)
			if err != nil {
				failures[i] = err.Error()
				continue
			}
			normalizedOp.Text = normalizedText
		}

		if op.CodeID != "" {
			if _, exists := proj.Codes[op.CodeID]; !exists {
				failures[i] = fmt.Sprintf("code not found: %s", op.CodeID)
				continue
			}
			normalizedOp.CodeID = op.CodeID
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
		Sections: normalizedSections,
		Failures: resultFailures,
	}, nil
}

func toUpdatedSectionsPayload(payload *file.UpdateCodeSectionsPayload) file.UpdatedCodeSectionsPayload {
	return file.UpdatedCodeSectionsPayload{
		Sections: payload.Sections,
		Failures: payload.Failures,
	}
}

func validateNoDuplicateSingleton(proj project.Project, payload file.CreateFilePayload, msg *commands.AnyMessage) error {
	if !payload.Type.IsSingleton() {
		return nil
	}
	if fileview.FindFileByType(proj, payload.Type) != nil {
		return utils.FieldError("type", "project already has a "+string(payload.Type))
	}
	return nil
}

func validateFileNotLocked(proj project.Project, payload file.ReplaceFileContentPayload, msg *commands.AnyMessage) error {
	return errIfLocked(proj, msg)
}

func errIfLocked(proj project.Project, msg *commands.AnyMessage) error {
	f := proj.Files[msg.AggregateID]
	if f.Locked {
		return utils.FieldError("file", "file is locked")
	}
	return nil
}

func errIfNotCorpus(proj project.Project, msg *commands.AnyMessage) error {
	f := proj.Files[msg.AggregateID]
	if f.Type != file.FileTypeCorpus {
		return utils.FieldError("file", "coding is only allowed on corpus files")
	}
	return nil
}

func validateCorpusFile[P any](proj project.Project, _ P, msg *commands.AnyMessage) error {
	return errIfNotCorpus(proj, msg)
}

func toReplacedContentPayload(payload *file.ReplaceFileContentPayload) file.ReplacedFileContentPayload {
	return file.ReplacedFileContentPayload{Content: payload.Content}
}

func transformEditFileContent(proj project.Project, payload file.EditFileContentPayload, msg *commands.AnyMessage) (file.ReplacedFileContentPayload, error) {
	if err := errIfLocked(proj, msg); err != nil {
		return file.ReplacedFileContentPayload{}, err
	}

	f := proj.Files[msg.AggregateID]
	if f.Content == "" {
		return file.ReplacedFileContentPayload{}, utils.FieldError("file", "has no content")
	}

	foundText, found := find.Find(payload.OldText, f.Content)
	if !found {
		return file.ReplacedFileContentPayload{}, utils.FieldError("old_text", "not found in file")
	}

	return file.ReplacedFileContentPayload{
		Content: find.Replace(f.Content, foundText, payload.NewText),
	}, nil
}
