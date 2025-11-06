package handlers

import (
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/project"
)

var Router = dispatch.LimitOnEntity(project.EntityName,
	dispatch.ToCreateEntityEvent[project.CreateProjectPayload](project.CreateProject, project.CreatedProject),
)
