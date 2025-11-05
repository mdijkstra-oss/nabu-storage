package projection

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"log/slog"
	"sync"
)

type ProjectView struct {
	ProjectStore *Store[project.Project]
	CodeStore    *Store[code.Code]
	FileStore    *Store[file.File]
}

// ApplyEventToAllStores Panics are caught and logged to prevent one store's failure from affecting others
func (pv *ProjectView) ApplyEventToAllStores(message *commands.AnyMessage) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic while applying event to stores",
				"action", message.Action,
				"aggregateID", message.AggregateID,
				"aggregateType", message.AggregateType,
				"panic", r)
		}
	}()

	pv.ProjectStore.ApplyEvent(message)
	pv.CodeStore.ApplyEvent(message)
	pv.FileStore.ApplyEvent(message)
}

type ProjectViewRegistry struct {
	projects       map[string]*ProjectView
	mu             sync.RWMutex
	projectReducer Reducer[project.Project]
	codeReducer    Reducer[code.Code]
	fileReducer    Reducer[file.File]
}

func NewProjectViewRegistry(
	projectReducer Reducer[project.Project],
	codeReducer Reducer[code.Code],
	fileReducer Reducer[file.File],
) *ProjectViewRegistry {
	return &ProjectViewRegistry{
		projects:       make(map[string]*ProjectView),
		projectReducer: projectReducer,
		codeReducer:    codeReducer,
		fileReducer:    fileReducer,
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

	view = pvr.createProjectView()
	pvr.projects[projectID] = view
	slog.Info("added new project to registry", "projectID", projectID)
	return view
}

func (pvr *ProjectViewRegistry) createProjectView() *ProjectView {
	return &ProjectView{
		ProjectStore: NewStore(pvr.projectReducer),
		CodeStore:    NewStore(pvr.codeReducer),
		FileStore:    NewStore(pvr.fileReducer),
	}
}
