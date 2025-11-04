package main

import (
	"github.com/go-chi/chi/v5"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	codeview "hermes-relay/internal/domain/projections/code-entity"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	"hermes-relay/internal/handlers"
	"hermes-relay/internal/lib/utils"
	"hermes-relay/internal/persistence"
	"hermes-relay/internal/projection"
	"log"
	"log/slog"
	net "net/http"
	"os"
)

func main() {

	// Todo: Change to env var
	setupLogger(slog.LevelDebug)

	var publisher = cqrs.NewInMemoryPublisher()

	//utils.MustNotError(PublishNewSourceFiles(publisher.Publish))

	registry := setupProjectViewRegistry(publisher)

	// All except views / projections must be after replay ⚠️
	utils.MustNotError(persistence.ReplayAllEvents(publisher))

	setUpCommandHandlers(publisher)

	publisher.Subscribe(cqrs.LimitOnType(cqrs.DomainEvent, cqrs.ReadOnlyRoutes(persistence.Apply)))

	slog.Info("Initializing command persistence")

	slog.Info("Initializing http endpoints")
	r := chi.NewRouter()
	handlers.SetupHTTPHandlers(r, publisher, registry)

	log.Fatal(net.ListenAndServe(":8080", r))
}

func setUpCommandHandlers(publisher *cqrs.InMemoryPublisher) {
	slog.Info("Setting up command handlers for new incoming messages")

	var commandRouter = cqrs.CombineRouters(
		code.Router,
		file.Router,
		project.Router,
	)

	publisher.Subscribe(commandRouter)
}

func setupProjectViewRegistry(publisher *cqrs.InMemoryPublisher) *projection.ProjectViewRegistry {
	registry := projection.NewProjectViewRegistry(
		projectview.Reducer,
		codeview.Reducer,
		fileview.Reducer,
	)

	publisher.Subscribe(cqrs.LimitOnType(cqrs.DomainEvent, cqrs.ReadOnlyRoutes(func(message *cqrs.AnyMessage) error {
		projectID := extractProjectID(message)
		if projectID == "" {
			return nil
		}

		projectRegistry := registry.EnsureProjectExists(message, projectID)
		if projectRegistry != nil {
			projectRegistry.ApplyEventToAllStores(message)
		}

		return nil
	})))

	return registry
}

func extractProjectID(message *cqrs.AnyMessage) string {
	if message.AggregateType == "Project" {
		return message.AggregateID
	}

	if payload, ok := message.Payload.(map[string]any); ok {
		if projectID, ok := payload["project_id"].(string); ok {
			return projectID
		}
	}

	return ""
}

func setupLogger(level slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
}
