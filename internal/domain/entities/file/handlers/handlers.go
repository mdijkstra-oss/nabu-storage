package handlers

import (
	"fmt"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/lib/text-search/find"
	"hermes-relay/internal/lib/utils"
	"strings"
)

const DEDUP_SIMILARITY_THRESHOLD = 0.8

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
				registry.TransformDomain[file.AddCodeSectionsPayload, file.ModifiedCodeSectionsPayload](
					reg,
					handleAddCodeSections,
					dispatch.ToUpdateEntityEvent[file.ModifiedCodeSectionsPayload, file.ModifiedCodeSectionsPayload](
						file.AddCodeSections,
						file.ModifiedCodeSections,
						assignOperationIDs,
					),
				),
			),

		dispatch.LimitOnAction(file.UpdateCodeSections,
			registry.TransformDomain[file.UpdateCodeSectionsPayload, file.ModifiedCodeSectionsPayload](
				reg,
				handleUpdateCodeSections,
				dispatch.ToUpdateEntityEvent[file.ModifiedCodeSectionsPayload, file.ModifiedCodeSectionsPayload](
					file.UpdateCodeSections,
					file.ModifiedCodeSections,
				),
			),
		),

		dispatch.LimitOnAction(file.RemoveCodeSections,
			registry.TransformDomain[file.RemoveCodeSectionsPayload, file.ModifiedCodeSectionsPayload](
				reg,
				handleRemoveCodeSections,
				dispatch.ToUpdateEntityEvent[file.ModifiedCodeSectionsPayload, file.ModifiedCodeSectionsPayload](
					file.RemoveCodeSections,
					file.ModifiedCodeSections,
				),
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

func textSimilarity(text1, text2 string) float64 {
	tokens1 := find.Tokenize(text1)
	tokens2 := find.Tokenize(text2)
	if len(tokens1) == 0 {
		return 0
	}
	return find.TokenOverlap(tokens1, tokens2)
}

func isSimilarSection(codeID, text string, existing file.CodedSection) bool {
	if codeID != existing.CodeID {
		return false
	}
	return textSimilarity(text, existing.Text) >= DEDUP_SIMILARITY_THRESHOLD
}

func findSimilarSection(codeID, text string, sections []file.CodedSection) *file.CodedSection {
	for i := range sections {
		if isSimilarSection(codeID, text, sections[i]) {
			return &sections[i]
		}
	}
	return nil
}

func isLLMActor(actorType commands.ActorType) bool {
	return actorType == commands.ActorTypeLLM
}

func validateLLMReasons(sections []file.AddSectionOp, actorType commands.ActorType) error {
	if !isLLMActor(actorType) {
		return nil
	}

	failures := make(map[int]string)
	for i, section := range sections {
		if section.Reason == "" {
			failures[i] = "reason is required for LLM actor"
		}
	}

	if len(failures) > 0 {
		return utils.ArrayItemErrors("sections", failures)
	}

	return nil
}

func mapAddToModified(payload file.AddCodeSectionsPayload) file.ModifiedCodeSectionsPayload {
	return file.ModifiedCodeSectionsPayload{
		Operations: utils.Map(payload.Sections, func(op file.AddSectionOp) file.SectionOp {
			return file.SectionOp{
				Op:         "add",
				CodeID:     op.CodeID,
				Text:       op.Text,
				Reason:     op.Reason,
				Confidence: &op.Confidence,
			}
		}),
	}
}

func mapUpdateToModified(payload file.UpdateCodeSectionsPayload) file.ModifiedCodeSectionsPayload {
	return file.ModifiedCodeSectionsPayload{
		Operations: utils.Map(payload.Sections, func(op file.UpdateSectionOp) file.SectionOp {
			return file.SectionOp{
				Op:         "update",
				ID:         op.ID,
				CodeID:     op.CodeID,
				Text:       op.Text,
				Reason:     op.Reason,
				Confidence: op.Confidence,
			}
		}),
	}
}

func mapRemoveToModified(payload file.RemoveCodeSectionsPayload) file.ModifiedCodeSectionsPayload {
	return file.ModifiedCodeSectionsPayload{
		Operations: utils.Map(payload.SectionIDs, func(id string) file.SectionOp {
			return file.SectionOp{
				Op: "delete",
				ID: id,
			}
		}),
	}
}

func handleAddCodeSections(proj project.Project, payload file.AddCodeSectionsPayload, msg *commands.AnyMessage) (file.ModifiedCodeSectionsPayload, error) {
	if err := validateLLMReasons(payload.Sections, msg.Actor.ActorType); err != nil {
		return file.ModifiedCodeSectionsPayload{}, err
	}
	return normalizeModifiedSections(proj, mapAddToModified(payload), msg)
}

func handleUpdateCodeSections(proj project.Project, payload file.UpdateCodeSectionsPayload, msg *commands.AnyMessage) (file.ModifiedCodeSectionsPayload, error) {
	return normalizeModifiedSections(proj, mapUpdateToModified(payload), msg)
}

func handleRemoveCodeSections(proj project.Project, payload file.RemoveCodeSectionsPayload, msg *commands.AnyMessage) (file.ModifiedCodeSectionsPayload, error) {
	return normalizeModifiedSections(proj, mapRemoveToModified(payload), msg)
}

type operationContext struct {
	fileContent  string
	fileCodes    []file.CodedSection
	projectCodes map[string]code.Code
}

type operationHandler func(file.SectionOp, operationContext) (file.SectionOp, error)

var operationHandlers = map[string]operationHandler{
	"add":    normalizeAddOp,
	"update": normalizeUpdateOp,
	"delete": normalizeDeleteOp,
}

func sectionExists(sectionID string, codes []file.CodedSection) bool {
	for _, section := range codes {
		if section.ID == sectionID {
			return true
		}
	}
	return false
}

func codeExists(codeID string, codes map[string]code.Code) bool {
	_, exists := codes[codeID]
	return exists
}

func validateCodeExists(codeID string, codes map[string]code.Code) error {
	if !codeExists(codeID, codes) {
		return fmt.Errorf("code not found")
	}
	return nil
}

func normalizeAddOp(op file.SectionOp, ctx operationContext) (file.SectionOp, error) {
	if err := validateCodeExists(op.CodeID, ctx.projectCodes); err != nil {
		return file.SectionOp{}, err
	}

	normalizedText, err := validateAndNormalizeText(op.Text, ctx.fileContent)
	if err != nil {
		return file.SectionOp{}, err
	}

	similarSection := findSimilarSection(op.CodeID, normalizedText, ctx.fileCodes)
	if similarSection != nil {
		updateOp := file.SectionOp{
			Op:         "update",
			ID:         similarSection.ID,
			CodeID:     op.CodeID,
			Text:       normalizedText,
			Reason:     op.Reason,
			Confidence: op.Confidence,
		}
		return normalizeUpdateOp(updateOp, ctx)
	}

	return file.SectionOp{
		Op:         "add",
		CodeID:     op.CodeID,
		Text:       normalizedText,
		Reason:     op.Reason,
		Confidence: op.Confidence,
	}, nil
}

func normalizeUpdateOp(op file.SectionOp, ctx operationContext) (file.SectionOp, error) {
	if !sectionExists(op.ID, ctx.fileCodes) {
		return file.SectionOp{}, fmt.Errorf("section not found")
	}

	if op.CodeID != "" {
		if err := validateCodeExists(op.CodeID, ctx.projectCodes); err != nil {
			return file.SectionOp{}, err
		}
	}

	if op.Text != "" {
		normalizedText, err := validateAndNormalizeText(op.Text, ctx.fileContent)
		if err != nil {
			return file.SectionOp{}, err
		}
		return file.SectionOp{
			Op:         "update",
			ID:         op.ID,
			CodeID:     op.CodeID,
			Text:       normalizedText,
			Reason:     op.Reason,
			Confidence: op.Confidence,
		}, nil
	}

	return op, nil
}

func normalizeDeleteOp(op file.SectionOp, ctx operationContext) (file.SectionOp, error) {
	return op, nil
}

func normalizeModifiedSections(proj project.Project, payload file.ModifiedCodeSectionsPayload, msg *commands.AnyMessage) (file.ModifiedCodeSectionsPayload, error) {
	if err := errIfNotCorpus(proj, msg); err != nil {
		return file.ModifiedCodeSectionsPayload{}, err
	}

	f := proj.Files[msg.AggregateID]
	ctx := operationContext{
		fileContent:  f.Content,
		fileCodes:    f.Codes,
		projectCodes: proj.Codes,
	}

	operations := []file.SectionOp{}
	failures := make(map[int]string)

	for i, op := range payload.Operations {
		handler, exists := operationHandlers[op.Op]
		if !exists {
			panic("unknown operation type: " + op.Op)
		}

		normalizedOp, err := handler(op, ctx)
		if err != nil {
			failures[i] = err.Error()
			continue
		}

		operations = append(operations, normalizedOp)
	}

	if len(failures) == len(payload.Operations) {
		return file.ModifiedCodeSectionsPayload{}, utils.ArrayItemErrors("operations", failures)
	}

	var resultFailures map[int]string
	if len(failures) > 0 {
		resultFailures = failures
	}

	return file.ModifiedCodeSectionsPayload{
		Operations: operations,
		Failures:   resultFailures,
	}, nil
}

func assignOperationIDs(payload *file.ModifiedCodeSectionsPayload) file.ModifiedCodeSectionsPayload {
	return file.ModifiedCodeSectionsPayload{
		Operations: utils.Map(payload.Operations, func(op file.SectionOp) file.SectionOp {
			if op.Op == "add" && op.ID == "" {
				op.ID = utils.NewID()
			}
			return op
		}),
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
