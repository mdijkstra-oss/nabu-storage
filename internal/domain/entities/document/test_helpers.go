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
		UpdatedAt:   th.DefaultTestTime(),
		Original:    "",
		Pinned:      false,
		Tags:        []string{},
		Content:     []Block{},
		Annotations: map[string]Annotation{},
	}
	merged := utils.ApplyPartialUpdate(defaults, overrides)
	return Document{
		ID:           id,
		Healthy:      true,
		Version:      1,
		DocumentData: merged,
	}
}
