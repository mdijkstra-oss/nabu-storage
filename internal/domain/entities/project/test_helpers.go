package project

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/lib/utils"
)

func BuildTestProject(id string, overrides ProjectData) Project {
	defaults := ProjectData{
		Name:        "Test Project",
		Description: "Test description",
	}
	merged := utils.ApplyPartialUpdate(defaults, overrides)
	return Project{
		ID:          id,
		Healthy:     true,
		ProjectData: merged,
		Codes:       make(map[string]code.Code),
		Files:       make(map[string]file.File),
	}
}
