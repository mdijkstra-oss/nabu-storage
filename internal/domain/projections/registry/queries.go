package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	"hermes-relay/internal/lib/utils"
)

func QueryAllProjects(state *RegistryState, query projection.PaginationQuery) []projection.PaginationResult[projectview.ProjectView] {
	allProjects := state.GetAllProjects()
	projectViews := utils.Map(allProjects, projectview.ToView)
	return projection.Paginate(projectViews, query)
}

func QueryProjectEvents(query projection.CursorQuery, projectID string, state *RegistryState) projection.CursorResult[commands.AnyMessage] {
	events := state.GetProjectEvents(projectID)
	return projection.CursorFilter(events, query, func(m commands.AnyMessage) string {
		return m.GetActorType()
	})
}
