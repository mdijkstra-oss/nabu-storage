package file

import (
	"hermes-relay/internal/cqrs/commands"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"hermes-relay/internal/lib/utils"
)

func BuildTestFile(id string, overrides FileData) File {
	defaults := FileData{
		ProjectID:   "project-1",
		Name:        "test.txt",
		Description: "",
		Title:       "",
		Summary:     "",
		Time:        th.DefaultTestTime(),
		Type:        FileTypeCorpus,
		Original:    "",
		Locked:      false,
	}
	merged := utils.ApplyPartialUpdate(defaults, overrides)
	return File{
		ID:       id,
		Healthy:  true,
		Version:  1,
		FileData: merged,
		Content:  "",
		Codes:    []CodedSection{},
	}
}

func BuildTestCodedSection(id, codeID, text string) CodedSection {
	return CodedSection{
		ID:         id,
		CodeID:     codeID,
		Text:       text,
		Confidence: ConfidenceHigh,
	}
}

func CreatedFileEventWithType(id, projectID, content string, fileType FileType) *commands.AnyMessage {
	return domain_helpers.NewDomainEvent(EntityName, id, CreatedFile, CreatedFilePayload{
		FileData: FileData{
			ProjectID: projectID,
			Name:      "test-file.txt",
			Type:      fileType,
			Locked:    fileType.IsLocked(),
		},
		Content: content,
		Codes:   []CodedSection{},
	})
}

func CreatedFileWithSectionsEvent(id, projectID, content string, sections []CodedSection) *commands.AnyMessage {
	return domain_helpers.NewDomainEvent(EntityName, id, CreatedFile, CreatedFilePayload{
		FileData: FileData{
			ProjectID: projectID,
			Name:      "test-file.txt",
			Type:      FileTypeCorpus,
			Locked:    true,
		},
		Content: content,
		Codes:   sections,
	})
}
