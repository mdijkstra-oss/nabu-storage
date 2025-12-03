package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities"
	codeview "hermes-relay/internal/domain/projections/code-entity"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	projectview "hermes-relay/internal/domain/projections/project-entity"
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
	//r.Post("/events", http.EventHandler(publisher.Publish))
	r.Get("/ws/{projectId}", websocket.Handler(hub, registryState, publisher.Subscribe))

	r.Route("/schemas", func(r chi.Router) {
		r.Get("/project", http.SchemaHandler(entities.ProjectSchema))
		r.Get("/file", http.SchemaHandler(entities.FileSchema))
		r.Get("/code", http.SchemaHandler(entities.CodeSchema))
	})

	// ⚠️ No joins / complex queries
	// That would probably mean that you'd have to reduce into a new entity
	r.Route("/queries/projects", func(r chi.Router) {
		// Todo: r.Use(middleware.RequireAuth)

		r.Get("/", http.Query(func(query projection.PaginationQuery) []projection.PaginationResult[projectview.ProjectView] {
			return registry.QueryAllProjects(registryState, query)
		}))

		r.Route("/{projectId}", func(r chi.Router) {
			// Todo: r.Use(middleware.RequireProjectAccess)

			r.Get("/", http.ProjectQuery(registryState, projectview.QueryProject))
			r.Get("/events", http.RegistryQuery(registryState, registry.QueryProjectEvents))
			r.Get("/search", http.ProjectQuery(registryState, fileview.SearchProject))

			r.Route("/files", func(r chi.Router) {
				r.Get("/", http.ProjectQuery(registryState, fileview.QueryFiles))
				r.Get("/{id}", http.ProjectQuery(registryState, fileview.QueryFile))
				r.Get("/{id}/chunks", http.ProjectQuery(registryState, fileview.QueryChunk))
			})

			r.Get("/codebook", http.ProjectQuery(registryState, fileview.QueryCodebook))

			r.Route("/codes", func(r chi.Router) {
				r.Get("/", http.ProjectQuery(registryState, codeview.QueryCodes))
				r.Get("/{id}", http.ProjectQuery(registryState, codeview.QueryCode))
				r.Get("/{id}/sections", http.ProjectQuery(registryState, codeview.QuerySections))
			})
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
