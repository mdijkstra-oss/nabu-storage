package projectview

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/project"
)

func QueryProject(query projection.EmptyQuery, proj project.Project) *project.Project {
	return &proj
}
