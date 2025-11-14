package http

import (
	"github.com/go-chi/chi/v5"
	"hermes-relay/internal/domain/projections/registry"
	"net/http"
)

func WithProject(registryState *registry.RegistryState) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			projectID := chi.URLParam(r, "projectId")

			if projectID == "" {
				http.Error(w, "projectId is required", http.StatusBadRequest)
				return
			}

			proj := registryState.GetProject(projectID)
			if proj == nil {
				http.Error(w, "project not found", http.StatusNotFound)
				return
			}

			if !proj.IsHealthy() {
				http.Error(w, "project is in unhealthy state due to corrupted data", http.StatusServiceUnavailable)
				return
			}

			ctx := withProjectContext(r.Context(), proj)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
