package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
)

type Registry struct {
	Projects        map[string]project.Project
	EntityToProject map[string]string
	Events          map[string][]commands.AnyMessage
}

func (r Registry) GetID() string {
	return "singleton"
}

type Store = projection.Store[Registry]

func NewStore() *Store {
	return projection.NewStore(EmptyRegistry(), Reducer)
}

func EmptyRegistry() *Registry {
	return &Registry{
		Projects:        make(map[string]project.Project),
		EntityToProject: make(map[string]string),
		Events:          make(map[string][]commands.AnyMessage),
	}
}

func GetProject(r *Registry, id string) *project.Project {
	if proj, ok := r.Projects[id]; ok {
		return &proj
	}
	return nil
}

func GetAllProjects(r *Registry) []project.Project {
	return utils.Values(r.Projects)
}

func GetProjectIDForEntity(r *Registry, aggregateID string) string {
	return r.EntityToProject[aggregateID]
}

func ResolveProjectID(r *Registry, message *commands.AnyMessage) string {
	projectID := commands.ExtractProjectID(message)
	if projectID == "" {
		projectID = GetProjectIDForEntity(r, message.AggregateID)
	}
	return projectID
}

func GetProjectEvents(r *Registry, projectID string) []commands.AnyMessage {
	return r.Events[projectID]
}
