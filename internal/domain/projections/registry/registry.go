package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/project"
	"sync"
)

type RegistryState struct {
	mu    sync.RWMutex
	state *Registry
}

func NewRegistryState() *RegistryState {
	return &RegistryState{
		state: &Registry{
			Projects:        make(map[string]project.Project),
			EntityToProject: make(map[string]string),
			Events:          make(map[string][]commands.AnyMessage),
		},
	}
}

func (rs *RegistryState) ApplyEvent(message *commands.AnyMessage) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.state = Reducer(rs.state, message)
}

func (rs *RegistryState) GetProject(projectID string) *project.Project {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	if proj, ok := rs.state.Projects[projectID]; ok {
		return &proj
	}
	return nil
}

func (rs *RegistryState) GetProjectIDForEntity(aggregateType commands.AggregateType, aggregateID string) string {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	key := string(aggregateType) + ":" + aggregateID
	return rs.state.EntityToProject[key]
}

func (rs *RegistryState) ResolveProjectID(message *commands.AnyMessage) string {
	projectID := commands.ExtractProjectID(message)
	if projectID == "" {
		projectID = rs.GetProjectIDForEntity(message.AggregateType, message.AggregateID)
	}
	return projectID
}

func (rs *RegistryState) GetProjectEvents(projectID string) []commands.AnyMessage {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return rs.state.Events[projectID]
}
