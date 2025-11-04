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

	setUpCommandHandlers(publisher)
	setupEventHandlers(publisher)

	//utils.MustNotError(PublishNewSourceFiles(publisher.Publish))

	registry := setupProjectViewRegistry(publisher)

	r := chi.NewRouter()
	handlers.SetupHTTPHandlers(r, publisher, registry)

	log.Fatal(net.ListenAndServe(":8080", r))
}

func setUpCommandHandlers(publisher *cqrs.InMemoryPublisher) {
	var commandRouter = cqrs.CombineRouters(
		code.Router,
		file.Router,
		project.Router,
	)

	publisher.Subscribe(commandRouter)
}

func setupEventHandlers(publisher *cqrs.InMemoryPublisher) {
	publisher.Subscribe(project.EventHandlers)

	utils.MustNotError(persistence.ReplayAllEvents(publisher))

	// Persist must be after replay ⚠️
	publisher.Subscribe(cqrs.LimitOnType(cqrs.DomainEvent, cqrs.ReadOnlyRoutes(persistence.Apply)))
}

func setupProjectViewRegistry(publisher *cqrs.InMemoryPublisher) *projection.ProjectViewRegistry {
	registry := projection.NewProjectViewRegistry()

	projectIDs, err := persistence.GetProjectIDs()
	utils.MustNotError(err)

	for _, projectID := range projectIDs {
		view := &projection.ProjectView{
			ProjectStore: projection.NewStore(projectview.Reducer),
			CodeStore:    projection.NewStore(codeview.Reducer),
			FileStore:    projection.NewStore(fileview.Reducer),
		}
		registry.AddProject(projectID, view)
	}

	publisher.Subscribe(cqrs.LimitOnType(cqrs.DomainEvent, cqrs.ReadOnlyRoutes(func(message *cqrs.AnyMessage) error {
		projectID := extractProjectID(message)
		if projectID == "" {
			return nil
		}

		projectRegistry := registry.GetProject(projectID)
		if projectRegistry == nil {
			return nil
		}

		projectRegistry.ApplyEventToAllStores(message)

		return nil
	})))

	slog.Info("initialized project view registry", "projectCount", len(projectIDs))
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
