package handlers

import (
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/registry"
	"hermes-relay/internal/domain/entities/project"
)

func NewRouter(_ *registry.ProjectViewRegistry) dispatch.CommandRouter {
	return dispatch.LimitOnEntity(project.EntityName,
		dispatch.ToCreateEntityEvent[project.CreateProjectPayload](project.CreateProject, project.CreatedProject),
	)
}
