package registry

import "hermes-relay/internal/domain/entities/project"

type Registry struct {
	Projects        map[string]project.Project
	EntityToProject map[string]string
}

func (r Registry) GetID() string {
	return "singleton"
}
