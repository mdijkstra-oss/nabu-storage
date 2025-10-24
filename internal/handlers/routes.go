package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/projections/code-entity"
	"hermes-relay/internal/domain/projections/file-entity"
	"hermes-relay/internal/handlers/http"
	"hermes-relay/internal/handlers/typed-query"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	net "net/http"
)

func SetupHTTPHandlers(r chi.Router, publisher *cqrs.InMemoryPublisher) {
	r.Use(middleware.Logger) // Todo: log level
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.StripSlashes)
	r.Use(CORS)
	r.Use(http.WithHeaders(http.DefaultHeaders))

	r.Post("/commands", http.CommandHandler(publisher))
	r.Post("/events", http.EventHandler(publisher))
	r.Get("/ws/", http.WebSocketHandler(publisher))

	r.Route("/queries", func(r chi.Router) {

		r.Route("/files", func(r chi.Router) {
			r.Get("/", typedquery.RouteWithMap(fileview.Store, typedquery.GetAll, func(f []file.File) []file.BaseFile {
				return utils.Map(f, func(f file.File) file.BaseFile {
					return f.BaseFile
				})
			}))

			r.Get("/{id}",
				http.WithHeaders(http.MarkDownHeaders)(
					typedquery.RouteWithMap(fileview.Store, typedquery.GetById, func(f *file.File) string {
						return f.Content
					}),
				).ServeHTTP,
			)

			r.Get("/{id}/chunks/{index}", typedquery.Route(fileview.Store, fileview.GetFileChunk))
		})

		r.Route("/codes", func(r chi.Router) {
			r.Get("/", typedquery.Route(codeview.Store, typedquery.GetAll))
			r.Get("/{id}", typedquery.Route(codeview.Store, typedquery.GetById))
			r.Get("/slug/{slug}", typedquery.Route(codeview.Store, codeview.GetBySlug))
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
