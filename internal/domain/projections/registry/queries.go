package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	"hermes-relay/internal/lib/utils"
)

func QueryAllProjects(store *Store, query projection.PaginationQuery) []projection.PaginationResult[projectview.ProjectSummary] {
	allProjects := projection.Read(store, GetAllProjects)
	projectSummaries := utils.Map(allProjects, projectview.ToSummary)
	return projection.Paginate(projectSummaries, query)
}

func QueryProjectEvents(query projection.EmptyQuery, projectID string, store *Store) []commands.AnyMessage {
	return projection.Read(store, func(r *Registry) []commands.AnyMessage {
		return GetProjectEvents(r, projectID)
	})
}
