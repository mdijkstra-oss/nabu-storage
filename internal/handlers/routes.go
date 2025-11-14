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

		r.Get("/", http.ToJSON(func(req *net.Request) ([]projection.PaginationResult[project.Project], error) {
			allProjects := registryState.GetAllProjects()
			return projection.Paginate(allProjects, projection.PaginationQuery{}), nil
		}))

		r.Route("/{projectId}", func(r chi.Router) {
			// Todo: r.Use(middleware.RequireProjectAccess)
			r.Use(http.WithProject(registryState))

			r.Get("/", http.ToJSON(func(r *net.Request) (*project.Project, error) {
				return http.ProjectFromRequest(r), nil
			}))

			r.Route("/files", func(r chi.Router) {
				r.Get("/", http.ToJSON(func(r *net.Request) ([]fileview.FileSummary, error) {
					files := http.FilesFromRequest(r)
					return utils.Map(files, fileview.ToSummary), nil
				}))
				r.Get("/{id}", http.ToJSON(func(req *net.Request) (*fileview.FileSummary, error) {
					fileID := chi.URLParam(req, "id")
					f := http.FileFromContext(req.Context(), fileID)
					if f == nil {
						return nil, nil
					}
					summary := fileview.ToSummary(*f)
					return &summary, nil
				}))
				r.Get("/{id}/chunks", http.ToJSON(func(req *net.Request) (*chunk.ChunkResult, error) {
					f := http.FileFromContext(req.Context(), chi.URLParam(req, "id"))
					if f == nil {
						return nil, nil
					}
					return chunk.GetChunk(*f, chunk.ChunkQuery{}), nil
				}))
			})

			r.Route("/codes", func(r chi.Router) {
				r.Get("/", http.ToJSON(func(req *net.Request) ([]projection.PaginationResult[code.Code], error) {
					codes := http.CodesFromRequest(req)
					return projection.Paginate(codes, projection.PaginationQuery{}), nil
				}))
				r.Get("/{id}", http.ToJSON(func(req *net.Request) (*code.Code, error) {
					codeID := chi.URLParam(req, "id")
					return http.CodeFromContext(req.Context(), codeID), nil
				}))
				r.Get("/slug/{slug}", http.ToJSON(func(req *net.Request) (*code.Code, error) {
					slug := chi.URLParam(req, "slug")
					proj := http.ProjectFromRequest(req)
					if proj == nil {
					return nil, nil
				}
				return projectview.GetCodeBySlug(*proj, slug), nil
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
