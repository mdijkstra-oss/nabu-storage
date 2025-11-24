package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"hermes-relay/internal/bootstrap"
	"hermes-relay/internal/config"
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
	cfg := config.Load()
	bootstrap.SetupLogger(cfg.LogLevel)

	var publisher = dispatch.NewInMemoryPublisher()
	hub := websocket.NewHub()

	registry := bootstrap.SetupRegistry(publisher, hub)

	// All except views / projections must be after replay ⚠️
	slog.Info("Initializing command persistence", "dir", cfg.PersistenceDir)
	disk := persistence.NewDiskPersistence(cfg.PersistenceDir)
	utils.MustNotError(disk.ReplayAllEvents(publisher))

	bootstrap.SetupCommandHandlers(publisher, registry)

	publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, dispatch.ReadOnlyRoutes(disk.Apply())))

	slog.Info("Initializing http endpoints")
	r := chi.NewRouter()
	handlers.SetupHTTPHandlers(r, publisher, registry, hub, cfg.CorsOrigins)

	addr := fmt.Sprintf(":%s", cfg.Port)
	slog.Info("Server starting", "port", cfg.Port, "cors_origins", cfg.CorsOrigins)
	log.Fatal(net.ListenAndServe(addr, r))
}
