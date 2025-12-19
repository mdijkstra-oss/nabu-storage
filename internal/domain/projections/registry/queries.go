package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	"hermes-relay/internal/lib/utils"
)

func QueryAllProjects(state *RegistryState, query projection.PaginationQuery) []projection.PaginationResult[projectview.ProjectSummary] {
	allProjects := state.GetAllProjects()
	projectSummaries := utils.Map(allProjects, projectview.ToSummary)
	return projection.Paginate(projectSummaries, query)
}

func QueryProjectEvents(query projection.EmptyQuery, projectID string, state *RegistryState) []commands.AnyMessage {
	events := state.GetProjectEvents(projectID)
	// TODO: paginate on this query
	return events
}
