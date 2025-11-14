package main

import (
	"github.com/go-chi/chi/v5"
	"hermes-relay/internal/bootstrap"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/persistence"
	"hermes-relay/internal/handlers"
	"hermes-relay/internal/handlers/http/websocket"
	"hermes-relay/internal/lib/utils"
	"log"
	"log/slog"
	net "net/http"
)

func main() {
	// Todo: Change to env var
	bootstrap.SetupLogger(slog.LevelDebug)

	var publisher = dispatch.NewInMemoryPublisher()
	hub := websocket.NewHub()

	registry := bootstrap.SetupRegistry(publisher, hub)

	// All except views / projections must be after replay ⚠️
	disk := persistence.New()
	utils.MustNotError(disk.ReplayAllEvents(publisher))

	bootstrap.SetupCommandHandlers(publisher, registry)

	publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, dispatch.ReadOnlyRoutes(disk.Apply())))

	slog.Info("Initializing command persistence")

	slog.Info("Initializing http endpoints")
	r := chi.NewRouter()
	handlers.SetupHTTPHandlers(r, publisher, registry, hub)

	log.Fatal(net.ListenAndServe(":8080", r))
}
