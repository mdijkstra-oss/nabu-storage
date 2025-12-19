package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/projections/project-entity"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/handlers/http"
	"hermes-relay/internal/handlers/http/websocket"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	net "net/http"
)

func SetupHTTPHandlers(r chi.Router, publisher *dispatch.InMemoryPublisher, registryState *registry.RegistryState, hub *websocket.Hub, corsOrigins []string) {
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.StripSlashes)
	r.Use(buildCorsHandler(corsOrigins))
	r.Use(http.WithHeaders(http.DefaultHeaders))

	r.Post("/commands", http.CommandHandler(publisher.Publish))
	r.Get("/ws/{projectId}", websocket.Handler(hub, registryState, publisher.Subscribe))

	r.Route("/queries/projects", func(r chi.Router) {
		r.Get("/", http.Query(func(query projection.PaginationQuery) []projection.PaginationResult[projectview.ProjectSummary] {
			return registry.QueryAllProjects(registryState, query)
		}))

		r.Route("/{projectId}", func(r chi.Router) {
			r.Get("/events", http.RegistryQuery(registryState, registry.QueryProjectEvents))
		})
	})

	utils.MustNotError(chi.Walk(r, func(method, route string, handler net.Handler, middlewares ...func(net.Handler) net.Handler) error {
		slog.Debug("registered route", "method", method, "route", route)
		return nil
	}))
}

func buildCorsHandler(allowedOrigins []string) func(net.Handler) net.Handler {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"*"}
	}

	return cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})
}
