package projection

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"sync"
)

// ProjectView holds all projection stores for a single project
type ProjectView struct {
	ProjectStore *Store[project.Project]
	CodeStore    *Store[code.Code]
	FileStore    *Store[file.File]
}

// ProjectViewRegistry manages views for all projects
type ProjectViewRegistry struct {
	projects map[string]*ProjectView
	mu       sync.RWMutex
}

func NewProjectViewRegistry() *ProjectViewRegistry {
	return &ProjectViewRegistry{
		projects: make(map[string]*ProjectView),
	}
}

// AddProject adds a project view to the registry
func (pvr *ProjectViewRegistry) AddProject(projectID string, view *ProjectView) {
	pvr.mu.Lock()
	defer pvr.mu.Unlock()
	pvr.projects[projectID] = view
}

// GetProject retrieves a project view from the registry
func (pvr *ProjectViewRegistry) GetProject(projectID string) *ProjectView {
	pvr.mu.RLock()
	defer pvr.mu.RUnlock()
	return pvr.projects[projectID]
}
