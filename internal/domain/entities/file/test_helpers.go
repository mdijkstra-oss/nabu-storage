package file

import (
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
		Type:        FileTypeSource,
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
