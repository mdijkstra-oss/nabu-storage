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
		// Entity-specific command handlers
		code.Router,
		file.Router,
		project.Router,
	)

	publisher.Subscribe(commandRouter)
}

func setupEventHandlers(publisher *cqrs.InMemoryPublisher) {
	// Cross-entity event handlers (publish derived events)
	publisher.Subscribe(project.EventHandlers)

	// Replay all persisted events on boot
	utils.MustNotError(persistence.ReplayAllEvents(publisher))

	// Persist must be after replay ⚠️
	publisher.Subscribe(cqrs.LimitOnType(cqrs.DomainEvent, cqrs.ReadOnlyRoutes(persistence.Apply)))
}

func setupProjectViewRegistry(publisher *cqrs.InMemoryPublisher) *projection.ProjectViewRegistry {
	registry := projection.NewProjectViewRegistry()

	// Get all project IDs from disk
	projectIDs, err := persistence.GetProjectIDs()
	utils.MustNotError(err)

	// Create empty stores for each project
	for _, projectID := range projectIDs {
		view := &projection.ProjectView{
			ProjectStore: projection.NewStore(projectview.Reducer),
			CodeStore:    projection.NewStore(codeview.Reducer),
			FileStore:    projection.NewStore(fileview.Reducer),
		}
		registry.AddProject(projectID, view)
	}

	// Subscribe router that routes domain events to correct project stores based on projectID
	publisher.Subscribe(cqrs.LimitOnType(cqrs.DomainEvent, cqrs.ReadOnlyRoutes(func(message *cqrs.AnyMessage) error {
		projectID := extractProjectID(message)
		if projectID == "" {
			return nil
		}

		view := registry.GetProject(projectID)
		if view == nil {
			return nil
		}

		// Apply to all stores - reducers will filter by entity type
		if err := view.ProjectStore.ApplyEvent(message); err != nil {
			return err
		}
		if err := view.CodeStore.ApplyEvent(message); err != nil {
			return err
		}
		if err := view.FileStore.ApplyEvent(message); err != nil {
			return err
		}

		return nil
	})))

	slog.Info("initialized project view registry", "projectCount", len(projectIDs))
	return registry
}

// extractProjectID extracts the project ID from an event
func extractProjectID(message *cqrs.AnyMessage) string {
	// For Project events, the aggregateID is the projectID
	if message.AggregateType == "Project" {
		return message.AggregateID
	}

	// For Code/File events, extract from payload
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
