package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/projection"
	domainprojection "hermes-relay/internal/cqrs/registry"
	"hermes-relay/internal/domain/entities/project"
	codeview "hermes-relay/internal/domain/projections/code-entity"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	"hermes-relay/internal/domain/projections/file-entity/chunk"
	"hermes-relay/internal/handlers/http"
	tq "hermes-relay/internal/handlers/http/typed-query"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	net "net/http"
)

func SetupHTTPHandlers(r chi.Router, publisher *dispatch.InMemoryPublisher, registry *domainprojection.ProjectViewRegistry) {
	r.Use(middleware.Logger) // Todo: log level
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.StripSlashes)
	r.Use(CORS)
	r.Use(http.WithHeaders(http.DefaultHeaders))

	r.Post("/commands", http.CommandHandler(publisher.Publish))
	//r.Post("/events", http.EventHandler(publisher.Publish))
	r.Get("/ws/", http.WebSocketHandler(publisher.Publish, publisher.Subscribe))

	// ⚠️ No joins / complex queries
	// That would probably mean that you'd have to reduce into a new entity
	r.Route("/queries/projects", func(r chi.Router) {
		// Todo: r.Use(middleware.RequireAuth)

		r.Get("/", tq.ToRoute(tq.Query(func(query projection.PaginationQuery) []projection.PaginationResult[project.Project] {
			allProjects := registry.GetAllProjectEntities()
			return projection.Paginate(allProjects, query)
		})))

		r.Route("/{projectId}", func(r chi.Router) {
			// Todo: r.Use(middleware.RequireProjectAccess)
			r.Use(http.WithProjectView(registry))

			r.Get("/", tq.QueryOneRoute(http.ProjectStoreFromRequest, projection.ByID))

			r.Route("/files", func(r chi.Router) {
				r.Get("/", tq.QueryRoute(http.FileStoreFromRequest, projection.ThenMap(projection.ByAll, fileview.ToSummary)))
				r.Get("/{id}", tq.QueryOneRoute(http.FileStoreFromRequest, projection.ThenMap(projection.ByID, fileview.ToSummary)))
				r.Get("/{id}/chunks", tq.QueryOneRoute(http.FileStoreFromRequest, chunk.ByChunk))
			})

			r.Route("/codes", func(r chi.Router) {
				r.Get("/", tq.QueryRoute(http.CodeStoreFromRequest, projection.Paginate))
				r.Get("/{id}", tq.QueryOneRoute(http.CodeStoreFromRequest, projection.ByID))
				r.Get("/slug/{slug}", tq.QueryOneRoute(http.CodeStoreFromRequest, codeview.BySlug))
			})
		})
	})

	utils.MustNotError(chi.Walk(r, func(method, route string, handler net.Handler, middlewares ...func(net.Handler) net.Handler) error {
		slog.Debug("registered route", "method", method, "route", route)
		return nil
	}))
}

func CORS(next net.Handler) net.Handler {
	return net.HandlerFunc(func(w net.ResponseWriter, r *net.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(net.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
