package handlers

import (
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/projections/registry"
)

func NewRouter(_ *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		dispatch.LimitOnEntity(project.EntityName,
			dispatch.ToCreateEntityEvent[project.CreateProjectPayload, project.CreatedProjectPayload](project.CreateProject, project.CreatedProject),
			dispatch.ToUpdateEntityEvent[project.UpdateProjectPayload, project.UpdatedProjectPayload](project.UpdateProject, project.UpdatedProject),
		),
		dispatch.ToEmptyDomainEvent(project.DeleteProject, project.DeletedProject),
	)
}
