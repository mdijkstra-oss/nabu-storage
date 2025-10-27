package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/projections/code-entity"
	"hermes-relay/internal/domain/projections/file-entity"
	"hermes-relay/internal/handlers/http"
	tq "hermes-relay/internal/handlers/http/typed-query"
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

	r.Post("/commands", http.CommandHandler(publisher.Publish))
	//r.Post("/events", http.EventHandler(publisher.Publish))
	r.Get("/ws/", http.WebSocketHandler(publisher.Publish, publisher.Subscribe))

	r.Route("/queries", func(r chi.Router) {

		r.Route("/files", func(r chi.Router) {
			r.Get("/", tq.QueryRouteWithMap(fileview.Store, tq.GetAll, func(f []file.File) []file.BaseFile {
				return utils.Map(f, func(f file.File) file.BaseFile {
					return f.BaseFile
				})
			}))

			r.Get("/{id}",
				http.WithHeaders(http.MarkDownHeaders)(
					tq.QueryRouteWithMap(fileview.Store, tq.GetById, func(f *file.File) string {
						return f.Content
					}),
				).ServeHTTP,
			)

			r.Get("/{id}/chunks/{index}", tq.QueryRoute(fileview.Store, fileview.GetFileChunk))
		})

		r.Route("/codes", func(r chi.Router) {
			r.Get("/", tq.QueryRoute(codeview.Store, tq.GetAll))
			r.Get("/{id}", tq.QueryRoute(codeview.Store, tq.GetById))
			r.Get("/slug/{slug}", tq.QueryRoute(codeview.Store, codeview.GetBySlug))
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
