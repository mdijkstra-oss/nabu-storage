package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
	"log/slog"
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

func (rs *RegistryState) GetAllProjects() []project.Project {
	rs.mu.RLock()
	defer rs.mu.RUnlock()
	return utils.Values(rs.state.Projects)
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

func Validate[P any](registry *RegistryState, validator func(project.Project, P, *commands.AnyMessage) error, handler dispatch.CommandRouter) dispatch.CommandRouter {
	return func(message *commands.AnyMessage, publisher dispatch.PublishFunc) (*commands.AnyMessage, error) {
		projectID := registry.ResolveProjectID(message)

		proj := registry.GetProject(projectID)

		if proj == nil {
			return nil, utils.FieldError("ProjectID", "not found")
		}

		if !proj.IsHealthy() {
			return nil, &utils.InternalError{Message: "project is in unhealthy state due to corrupted data, commands are blocked"}
		}

		var payload P
		if err := commands.UnmarshallPayload(message, &payload); err != nil {
			slog.Warn("failed to unmarshal command payload, ignoring invalid request",
				"action", message.Action,
				"aggregateType", message.AggregateType,
				"error", err)
			return nil, utils.FieldError("payload", "invalid format")
		}

		validationErr := validator(*proj, payload, message)
		if validationErr != nil {
			return nil, validationErr
		}

		return handler(message, publisher)
	}
}
