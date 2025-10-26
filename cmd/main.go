package main

import (
	"github.com/go-chi/chi/v5"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/projections/code-entity"
	"hermes-relay/internal/domain/projections/file-entity"
	"hermes-relay/internal/handlers"
	"hermes-relay/internal/lib/utils"
	"hermes-relay/internal/persistence"
	"log"
	"log/slog"
	net "net/http"
	"os"
)

func main() {

	// Todo: Change to env var
	setupLogger(slog.LevelDebug)

	var publisher = cqrs.NewInMemoryPublisher()

	setUpCommandHandlers(publisher)
	setupEventHandlers(publisher)

	//utils.MustNotError(PublishNewSourceFiles(publisher.Publish))

	r := chi.NewRouter()
	handlers.SetupHTTPHandlers(r, publisher)

	log.Fatal(net.ListenAndServe(":8080", r))
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

func setupLogger(level slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
}
