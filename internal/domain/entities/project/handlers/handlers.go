package handlers

import (
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/projections/registry"
)

func NewRouter(_ *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		dispatch.LimitOnEntity(project.EntityName,
			dispatch.ToCreateEntityEvent[project.CreateProjectPayload, project.CreatedProjectPayload](project.CreateProject, project.CreatedProject, createProjectFromPayload),
			dispatch.ToUpdateEntityEvent[project.UpdateProjectPayload, project.UpdatedProjectPayload](project.UpdateProject, project.UpdatedProject),
			dispatch.ToEmptyDomainEvent(project.PinProject, project.PinnedProject),
			dispatch.ToEmptyDomainEvent(project.UnpinProject, project.UnpinnedProject),
			dispatch.ToUpdateEntityEvent[project.ChangePhasePayload, project.ChangedPhasePayload](project.ChangePhase, project.ChangedPhase),
		),
		dispatch.ToEmptyDomainEvent(project.DeleteProject, project.DeletedProject),
	)
}

func createProjectFromPayload(payload *project.CreateProjectPayload) project.CreatedProjectPayload {
	phase := payload.Phase
	if phase == "" {
		phase = project.PhaseExplore
	}
	return project.ProjectData{
		Name:        payload.Name,
		Description: payload.Description,
		Phase:       phase,
	}
}
