package projection

import (
	"hermes-relay/internal/cqrs"
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
func (pv *ProjectView) ApplyEventToAllStores(message *cqrs.AnyMessage) {
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
	projects map[string]*ProjectView
	mu       sync.RWMutex
}

func NewProjectViewRegistry() *ProjectViewRegistry {
	return &ProjectViewRegistry{
		projects: make(map[string]*ProjectView),
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
