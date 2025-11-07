package http

import (
	"github.com/go-chi/chi/v5"
	domainprojection "hermes-relay/internal/cqrs/registry"
	"net/http"
)

func WithProjectView(registry *domainprojection.ProjectViewRegistry) func(http.Handler) http.Handler {
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

			if !view.IsHealthy() {
				http.Error(w, "project is in unhealthy state due to corrupted data", http.StatusServiceUnavailable)
				return
			}

			ctx := withProjectViewContext(r.Context(), view)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
