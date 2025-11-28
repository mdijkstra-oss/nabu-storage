package file

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	th "hermes-relay/internal/lib/test-helpers"
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
		FileData: merged,
		Chunks:   []Chunk{},
	}
}

func BuildTestChunk(id, content string, codes []CodedSection) Chunk {
	return Chunk{
		ID:      id,
		Content: content,
		Codes:   codes,
	}
}

func BuildTestCodedSection(id, codeID, text string) CodedSection {
	return CodedSection{
		ID:     id,
		CodeID: codeID,
		Text:   text,
	}
}

func CreatedFileEvent(id, projectID, content string) *commands.AnyMessage {
	return domain_helpers.NewDomainEvent(EntityName, id, CreatedFile, CreatedFilePayload{
		FileData: FileData{
			ProjectID: projectID,
			Name:      "test-file.txt",
			Type:      FileTypeCorpus,
			Locked:    true,
		},
		Chunks: []Chunk{
			{ID: "chunk-1", Content: content, Codes: []CodedSection{}},
		},
	})
}
