package projection

import (
	"context"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"net/http"
)

type contextKey string

const projectViewKey contextKey = "projectView"

// WithProjectView adds a ProjectView to the context
func WithProjectView(ctx context.Context, view *ProjectView) context.Context {
	return context.WithValue(ctx, projectViewKey, view)
}

// ProjectViewFromContext retrieves the ProjectView from context
func ProjectViewFromContext(ctx context.Context) *ProjectView {
	view, _ := ctx.Value(projectViewKey).(*ProjectView)
	return view
}

// CodeStoreFromContext retrieves the code store from context
func CodeStoreFromContext(ctx context.Context) *Store[code.Code] {
	view := ProjectViewFromContext(ctx)
	if view == nil {
		return nil
	}
	return view.CodeStore
}

// FileStoreFromContext retrieves the file store from context
func FileStoreFromContext(ctx context.Context) *Store[file.File] {
	view := ProjectViewFromContext(ctx)
	if view == nil {
		return nil
	}
	return view.FileStore
}

// ProjectStoreFromContext retrieves the project store from context
func ProjectStoreFromContext(ctx context.Context) *Store[project.Project] {
	view := ProjectViewFromContext(ctx)
	if view == nil {
		return nil
	}
	return view.ProjectStore
}

// CodeStoreForProject retrieves the code store for a specific project
func CodeStoreForProject(registry *ProjectViewRegistry, projectID string) *Store[code.Code] {
	view := registry.GetProject(projectID)
	if view == nil {
		return nil
	}
	return view.CodeStore
}

// FileStoreForProject retrieves the file store for a specific project
func FileStoreForProject(registry *ProjectViewRegistry, projectID string) *Store[file.File] {
	view := registry.GetProject(projectID)
	if view == nil {
		return nil
	}
	return view.FileStore
}

// ProjectStoreForProject retrieves the project store for a specific project
func ProjectStoreForProject(registry *ProjectViewRegistry, projectID string) *Store[project.Project] {
	view := registry.GetProject(projectID)
	if view == nil {
		return nil
	}
	return view.ProjectStore
}

// CodeStoreFromRequest retrieves the code store from http request context
func CodeStoreFromRequest(r *http.Request) *Store[code.Code] {
	return CodeStoreFromContext(r.Context())
}

// FileStoreFromRequest retrieves the file store from http request context
func FileStoreFromRequest(r *http.Request) *Store[file.File] {
	return FileStoreFromContext(r.Context())
}

// ProjectStoreFromRequest retrieves the project store from http request context
func ProjectStoreFromRequest(r *http.Request) *Store[project.Project] {
	return ProjectStoreFromContext(r.Context())
}
