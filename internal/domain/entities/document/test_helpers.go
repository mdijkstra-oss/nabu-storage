package document

import (
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/utils"
)

func BuildTestDocument(id string, overrides DocumentData) Document {
	defaults := DocumentData{
		ProjectID:   "project-1",
		Name:        "test.txt",
		Description: "",
		Title:       "",
		Time:        th.DefaultTestTime(),
		Original:    "",
		Pinned:      false,
		Content:     []Block{},
		Annotations: nil,
	}
	merged := utils.ApplyPartialUpdate(defaults, overrides)
	return Document{
		ID:           id,
		Healthy:      true,
		Version:      1,
		DocumentData: merged,
	}
}
