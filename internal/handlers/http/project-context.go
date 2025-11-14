package http

import (
	"context"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	"net/http"
)

type contextKey string

const projectKey contextKey = "project"

func withProjectContext(ctx context.Context, proj *project.Project) context.Context {
	return context.WithValue(ctx, projectKey, proj)
}

func ProjectFromContext(ctx context.Context) *project.Project {
	proj, _ := ctx.Value(projectKey).(*project.Project)
	return proj
}

func CodesFromContext(ctx context.Context) []code.Code {
	proj := ProjectFromContext(ctx)
	if proj == nil {
		return nil
	}
	return projectview.GetAllCodes(*proj)
}

func FilesFromContext(ctx context.Context) []file.File {
	proj := ProjectFromContext(ctx)
	if proj == nil {
		return nil
	}
	return projectview.GetAllFiles(*proj)
}

func CodeFromContext(ctx context.Context, codeID string) *code.Code {
	proj := ProjectFromContext(ctx)
	if proj == nil {
		return nil
	}
	return projectview.GetCode(*proj, codeID)
}

func FileFromContext(ctx context.Context, fileID string) *file.File {
	proj := ProjectFromContext(ctx)
	if proj == nil {
		return nil
	}
	return projectview.GetFile(*proj, fileID)
}

func ProjectFromRequest(r *http.Request) *project.Project {
	return ProjectFromContext(r.Context())
}

func CodesFromRequest(r *http.Request) []code.Code {
	return CodesFromContext(r.Context())
}

func FilesFromRequest(r *http.Request) []file.File {
	return FilesFromContext(r.Context())
}
