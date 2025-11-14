package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/project"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	"hermes-relay/internal/domain/projections/file-entity/chunk"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/handlers/http"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	net "net/http"
)

func SetupHTTPHandlers(r chi.Router, publisher *dispatch.InMemoryPublisher, registryState *registry.RegistryState) {
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

		r.Get("/", http.Query(func(query projection.PaginationQuery) []projection.PaginationResult[project.Project] {
			allProjects := registryState.GetAllProjects()
			return projection.Paginate(allProjects, query)
		}))

		r.Route("/{projectId}", func(r chi.Router) {
			// Todo: r.Use(middleware.RequireProjectAccess)

			r.Get("/", http.ProjectQuery(registryState, func(query http.EmptyQuery, proj project.Project) *project.Project {
				return &proj
			}))

			r.Route("/files", func(r chi.Router) {
				r.Get("/", http.ProjectQuery(registryState, func(query projection.PaginationQuery, proj project.Project) []fileview.FileSummary {
					files := projectview.GetAllFiles(proj)
					return utils.Map(files, fileview.ToSummary)
				}))

				r.Get("/{id}", http.ProjectQuery(registryState, func(query http.IDQuery, proj project.Project) *fileview.FileSummary {
					f := projectview.GetFile(proj, query.ID)
					if f == nil {
						return nil
					}
					summary := fileview.ToSummary(*f)
					return &summary
				}))

				r.Get("/{id}/chunks", http.ProjectQuery(registryState, func(query chunk.ChunkQuery, proj project.Project) *chunk.ChunkResult {
					f := projectview.GetFile(proj, query.ID)
					if f == nil {
						return nil
					}
					return chunk.GetChunk(*f, query)
				}))
			})

			r.Route("/codes", func(r chi.Router) {
				r.Get("/", http.ProjectQuery(registryState, func(query projection.PaginationQuery, proj project.Project) []projection.PaginationResult[code.Code] {
					codes := projectview.GetAllCodes(proj)
					return projection.Paginate(codes, query)
				}))

				r.Get("/{id}", http.ProjectQuery(registryState, func(query http.IDQuery, proj project.Project) *code.Code {
					return projectview.GetCode(proj, query.ID)
				}))

				r.Get("/slug/{slug}", http.ProjectQuery(registryState, func(query http.SlugQuery, proj project.Project) *code.Code {
					return projectview.GetCodeBySlug(proj, query.Slug)
				}))
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
