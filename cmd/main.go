package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/projections/code-entity"
	"hermes-relay/internal/domain/projections/file-entity"
	"hermes-relay/internal/handlers"
	"hermes-relay/internal/handlers/typed-query"
	"hermes-relay/internal/lib/utils"
	"hermes-relay/internal/persistence"
	"log"
	"log/slog"
	"net/http"
	"os"
)

func main() {

	// Todo: Change to env var
	setupLogger(slog.LevelDebug)

	var publisher = cqrs.NewInMemoryPublisher()

	setUpCommandHandlers(publisher)
	setupEventHandlers(publisher)

	//utils.MustNotError(PublishNewSourceFiles(publisher.Publish, fileview.Store))

	setupHTTPHandlers(publisher)
}

func setUpCommandHandlers(publisher *cqrs.InMemoryPublisher) {
	var commandRouter = cqrs.CombineRouters(
		// Entity-specific command handlers
		code.Router,
		file.Router,
	)

	publisher.Subscribe(commandRouter)
}

func setupEventHandlers(publisher *cqrs.InMemoryPublisher) {
	// Domain event handlers (readonly routes)
	publisher.Subscribe(cqrs.LimitOnType(cqrs.DomainEvent, cqrs.ReadOnlyRoutes(
		fileview.Store.ApplyEvent,
		codeview.Store.ApplyEvent,
	)))

	// Replay all persisted events on boot
	utils.MustNotError(persistence.ReplayAllEvents(publisher))

	// Persist must be after replay ⚠️
	publisher.Subscribe(cqrs.LimitOnType(cqrs.DomainEvent, cqrs.ReadOnlyRoutes(persistence.Apply)))
}

func setupHTTPHandlers(publisher *cqrs.InMemoryPublisher) {
	r := chi.NewRouter()
	r.Use(middleware.Logger) // Todo: log level
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.StripSlashes)

	r.Post("/commands", handlers.CommandHandler(publisher))
	r.Post("/events", handlers.EventHandler(publisher))
	r.Get("/ws/", handlers.WebSocketHandler(publisher))

	r.Route("/queries", func(r chi.Router) {

		r.Route("/files", func(r chi.Router) {
			r.Get("/", typedquery.Route(fileview.Store, typedquery.GetAll))
			r.Get("/{id}", typedquery.Route(fileview.Store, typedquery.GetById))
			r.Get("/{id}/chunks/{index}", typedquery.Route(fileview.Store, fileview.GetFileChunk))
		})

		r.Route("/codes", func(r chi.Router) {
			r.Get("/", typedquery.Route(codeview.Store, typedquery.GetAll))
			r.Get("/{id}", typedquery.Route(codeview.Store, typedquery.GetById))
			r.Get("/s/{slug}", typedquery.Route(codeview.Store, codeview.GetBySlug))
		})

	})

	utils.MustNotError(chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		slog.Debug("registered route", "method", method, "route", route)
		return nil
	}))

	log.Fatal(http.ListenAndServe(":8080", CORS(r)))
}

func setupLogger(level slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
}

// Todo: specify who can post etc...

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
