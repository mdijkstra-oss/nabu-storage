package registry

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/project"
)

func QueryAllProjects(state *RegistryState, query projection.PaginationQuery) []projection.PaginationResult[project.Project] {
	allProjects := state.GetAllProjects()
	return projection.Paginate(allProjects, query)
}
