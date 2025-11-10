package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	"sync"
)

type ProjectView struct {
	ProjectID    string
	ProjectStore *projection.Store[project.Project]
	CodeStore    *projection.Store[code.Code]
	FileStore    *projection.Store[file.File]
	healthy      bool
	mu           sync.RWMutex
}

func (pv *ProjectView) IsHealthy() bool {
	pv.mu.RLock()
	defer pv.mu.RUnlock()
	return pv.healthy
}

func (pv *ProjectView) markUnhealthy() {
	pv.mu.Lock()
	defer pv.mu.Unlock()
	pv.healthy = false
}

func (pv *ProjectView) ApplyEventToAllStores(message *commands.AnyMessage) {
	if !pv.IsHealthy() {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			slog.Error("FATAL: corrupt event crashed projection, marking project unhealthy",
				"projectID", pv.ProjectID,
				"action", message.Action,
				"aggregateID", message.AggregateID,
				"aggregateType", message.AggregateType,
				"timestamp", message.Timestamp,
				"panic", r)
			pv.markUnhealthy()
		}
	}()

	pv.ProjectStore.ApplyEvent(message)
	pv.CodeStore.ApplyEvent(message)
	pv.FileStore.ApplyEvent(message)
}

type ProjectViewRegistry struct {
	projects       map[string]*ProjectView
	mu             sync.RWMutex
	projectReducer projection.Reducer[project.Project]
	codeReducer    projection.Reducer[code.Code]
	fileReducer    projection.Reducer[file.File]

	// Lookup table to map entity IDs back to project IDs
	// Key format: "AggregateType:AggregateID" -> projectID
	entityToProject map[string]string
}

func NewProjectViewRegistry(
	projectReducer projection.Reducer[project.Project],
	codeReducer projection.Reducer[code.Code],
	fileReducer projection.Reducer[file.File],
) *ProjectViewRegistry {
	return &ProjectViewRegistry{
		projects:        make(map[string]*ProjectView),
		projectReducer:  projectReducer,
		codeReducer:     codeReducer,
		fileReducer:     fileReducer,
		entityToProject: make(map[string]string),
	}
}

func (pvr *ProjectViewRegistry) AddProject(projectID string, view *ProjectView) {
	pvr.mu.Lock()
	defer pvr.mu.Unlock()
	pvr.projects[projectID] = view
}

func (pvr *ProjectViewRegistry) GetProject(projectID string) *ProjectView {
	pvr.mu.RLock()
	defer pvr.mu.RUnlock()
	return pvr.projects[projectID]
}

func (pvr *ProjectViewRegistry) RemoveProject(projectID string) {
	pvr.mu.Lock()
	defer pvr.mu.Unlock()
	delete(pvr.projects, projectID)
}

func (pvr *ProjectViewRegistry) GetProjectIDForEntity(aggregateType commands.AggregateType, aggregateID string) string {
	pvr.mu.RLock()
	defer pvr.mu.RUnlock()

	key := string(aggregateType) + ":" + aggregateID
	return pvr.entityToProject[key]
}

func (pvr *ProjectViewRegistry) GetAllProjectEntities() []project.Project {
	pvr.mu.RLock()
	defer pvr.mu.RUnlock()

	var result []project.Project
	for _, view := range pvr.projects {
		entities := view.ProjectStore.GetAll()
		result = append(result, entities...)
	}
	return result
}

func (pvr *ProjectViewRegistry) EnsureProjectExists(message *commands.AnyMessage, projectID string) *ProjectView {
	view := pvr.GetProject(projectID)
	if view != nil {
		return view
	}

	if message.AggregateType != "Project" || message.Action != "CreatedProject" {
		return nil
	}

	pvr.mu.Lock()
	defer pvr.mu.Unlock()

	if existing := pvr.projects[projectID]; existing != nil {
		return existing
	}

	view = pvr.createProjectView(projectID)
	pvr.projects[projectID] = view
	slog.Info("added new project to registry", "projectID", projectID)
	return view
}

func (pvr *ProjectViewRegistry) createProjectView(projectID string) *ProjectView {
	return &ProjectView{
		ProjectID:    projectID,
		ProjectStore: projection.NewStore(pvr.projectReducer),
		CodeStore:    projection.NewStore(pvr.codeReducer),
		FileStore:    projection.NewStore(pvr.fileReducer),
		healthy:      true,
	}
}

// UpdateEntityLookups updates the entity-to-project lookup table based on events
func (pvr *ProjectViewRegistry) UpdateEntityLookups(message *commands.AnyMessage, projectID string) {
	pvr.mu.Lock()
	defer pvr.mu.Unlock()

	key := string(message.AggregateType) + ":" + message.AggregateID

	if commands.IsCreatedEvent(message.Action) {
		pvr.entityToProject[key] = projectID
	} else if commands.IsDeletedEvent(message.Action) {
		delete(pvr.entityToProject, key)
	}
}

func Validate[P any](registry *ProjectViewRegistry, validator func(*ProjectView, P, *commands.AnyMessage) error, handler dispatch.CommandRouter) dispatch.CommandRouter {
	return func(message *commands.AnyMessage, publisher dispatch.PublishFunc) (*commands.AnyMessage, error) {
		projectId := commands.ExtractProjectID(message)

		// If ExtractProjectID didn't find a projectID, try looking it up by entity ChunkID
		if projectId == "" {
			projectId = registry.GetProjectIDForEntity(message.AggregateType, message.AggregateID)
		}

		view := registry.GetProject(projectId)

		if view == nil {
			return nil, utils.FieldError("ProjectID", "not found")
		}

		if !view.IsHealthy() {
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

		validationErr := validator(view, payload, message)
		if validationErr != nil {
			return nil, validationErr
		}

		return handler(message, publisher)
	}
}
