package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/project"
)

type Registry struct {
	Projects        map[string]project.Project
	EntityToProject map[string]string
	Events          map[string][]commands.AnyMessage
}

func (r Registry) GetID() string {
	return "singleton"
}
