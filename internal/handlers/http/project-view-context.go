package http

import (
	"context"
	"hermes-relay/internal/cqrs/projection"
	domainprojection "hermes-relay/internal/cqrs/registry"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"net/http"
)

type contextKey string

const projectViewKey contextKey = "projectView"

func withProjectViewContext(ctx context.Context, view *domainprojection.ProjectView) context.Context {
	return context.WithValue(ctx, projectViewKey, view)
}

func ProjectViewFromContext(ctx context.Context) *domainprojection.ProjectView {
	view, _ := ctx.Value(projectViewKey).(*domainprojection.ProjectView)
	return view
}

func CodeStoreFromContext(ctx context.Context) *projection.Store[code.Code] {
	view := ProjectViewFromContext(ctx)
	if view == nil {
		return nil
	}
	return view.CodeStore
}

func FileStoreFromContext(ctx context.Context) *projection.Store[file.File] {
	view := ProjectViewFromContext(ctx)
	if view == nil {
		return nil
	}
	return view.FileStore
}

func ProjectStoreFromContext(ctx context.Context) *projection.Store[project.Project] {
	view := ProjectViewFromContext(ctx)
	if view == nil {
		return nil
	}
	return view.ProjectStore
}

func CodeStoreFromRequest(r *http.Request) *projection.Store[code.Code] {
	return CodeStoreFromContext(r.Context())
}

func FileStoreFromRequest(r *http.Request) *projection.Store[file.File] {
	return FileStoreFromContext(r.Context())
}

func ProjectStoreFromRequest(r *http.Request) *projection.Store[project.Project] {
	return ProjectStoreFromContext(r.Context())
}
