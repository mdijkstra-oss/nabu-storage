package main

import (
	"github.com/go-chi/chi/v5"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/persistence"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	codeview "hermes-relay/internal/domain/projections/code-entity"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	"hermes-relay/internal/handlers"
	"hermes-relay/internal/lib/utils"
	"log"
	"log/slog"
	net "net/http"
	"os"
)

func main() {

	// Todo: Change to env var
	setupLogger(slog.LevelDebug)

	var publisher = dispatch.NewInMemoryPublisher()

	//utils.MustNotError(PublishNewSourceFiles(publisher.Publish))

	registry := setupProjectViewRegistry(publisher)

	// All except views / projections must be after replay ⚠️
	utils.MustNotError(persistence.ReplayAllEvents(publisher))

	setUpCommandHandlers(publisher)

	publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, dispatch.ReadOnlyRoutes(persistence.Apply)))

	slog.Info("Initializing command persistence")

	slog.Info("Initializing http endpoints")
	r := chi.NewRouter()
	handlers.SetupHTTPHandlers(r, publisher, registry)

	log.Fatal(net.ListenAndServe(":8080", r))
}

func setUpCommandHandlers(publisher *dispatch.InMemoryPublisher) {
	slog.Info("Setting up command handlers for new incoming messages")

	var commandRouter = dispatch.CombineRouters(
		code.Router,
		file.Router,
		project.Router,
	)

	publisher.Subscribe(commandRouter)
}

func setupProjectViewRegistry(publisher *dispatch.InMemoryPublisher) *projection.ProjectViewRegistry {
	registry := projection.NewProjectViewRegistry(
		projectview.Reducer,
		codeview.Reducer,
		fileview.Reducer,
	)

	publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, dispatch.ReadOnlyRoutes(func(message *commands.AnyMessage) error {
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

func extractProjectID(message *commands.AnyMessage) string {
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
