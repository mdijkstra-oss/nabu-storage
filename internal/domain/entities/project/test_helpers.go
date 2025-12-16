package project

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
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
		Version:     1,
		ProjectData: merged,
		Codes:       make(map[string]code.Code),
		Files:       make(map[string]file.File),
	}
}

func CreatedProjectEvent(id string) *commands.AnyMessage {
	return domain_helpers.NewDomainEvent(EntityName, id, CreatedProject, CreatedProjectPayload{
		Name: "Test Project",
	})
}
