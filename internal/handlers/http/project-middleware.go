package http

import (
	"github.com/go-chi/chi/v5"
	"hermes-relay/internal/cqrs/projection"
	"net/http"
)

func WithProjectView(registry *projection.ProjectViewRegistry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			projectID := chi.URLParam(r, "projectId")

			if projectID == "" {
				http.Error(w, "projectId is required", http.StatusBadRequest)
				return
			}

			view := registry.GetProject(projectID)
			if view == nil {
				http.Error(w, "project not found", http.StatusNotFound)
				return
			}

			ctx := projection.WithProjectView(r.Context(), view)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
