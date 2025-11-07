package main

import (
	"github.com/go-chi/chi/v5"
	"hermes-relay/internal/bootstrap"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/persistence"
	"hermes-relay/internal/handlers"
	"hermes-relay/internal/lib/utils"
	"log"
	"log/slog"
	net "net/http"
)

func main() {
	// Todo: Change to env var
	bootstrap.SetupLogger(slog.LevelDebug)

	var publisher = dispatch.NewInMemoryPublisher()

	registry := bootstrap.SetupProjectViewRegistry(publisher)

	// All except views / projections must be after replay ⚠️
	utils.MustNotError(persistence.ReplayAllEvents(publisher))

	bootstrap.SetupCommandHandlers(publisher, registry)

	publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, dispatch.ReadOnlyRoutes(persistence.Apply)))

	slog.Info("Initializing command persistence")

	slog.Info("Initializing http endpoints")
	r := chi.NewRouter()
	handlers.SetupHTTPHandlers(r, publisher, registry)

	log.Fatal(net.ListenAndServe(":8080", r))
}
