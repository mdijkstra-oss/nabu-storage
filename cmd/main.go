package main

import (
	"github.com/go-chi/chi/v5"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/persistence"
	domainprojection "hermes-relay/internal/cqrs/registry"
	codehandlers "hermes-relay/internal/domain/entities/code/handlers"
	filehandlers "hermes-relay/internal/domain/entities/file/handlers"
	projecthandlers "hermes-relay/internal/domain/entities/project/handlers"
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

	setUpCommandHandlers(publisher, registry)

	publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, dispatch.ReadOnlyRoutes(persistence.Apply)))

	slog.Info("Initializing command persistence")

	slog.Info("Initializing http endpoints")
	r := chi.NewRouter()
	handlers.SetupHTTPHandlers(r, publisher, registry)

	log.Fatal(net.ListenAndServe(":8080", r))
}

func setUpCommandHandlers(publisher *dispatch.InMemoryPublisher, registry *domainprojection.ProjectViewRegistry) {
	slog.Info("Setting up command handlers for new incoming messages")

	var commandRouter = dispatch.LimitOnType(commands.Command,
		dispatch.CombineRouters(
			codehandlers.NewRouter(registry),
			filehandlers.NewRouter(registry),
			projecthandlers.NewRouter(registry),
		),
	)

	publisher.Subscribe(commandRouter)
}

func setupProjectViewRegistry(publisher *dispatch.InMemoryPublisher) *domainprojection.ProjectViewRegistry {
	registry := domainprojection.NewProjectViewRegistry(
		projectview.Reducer,
		codeview.Reducer,
		fileview.Reducer,
	)

	publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, dispatch.ReadOnlyRoutes(func(message *commands.AnyMessage) error {
		projectID := commands.ExtractProjectID(message)
		if projectID == "" {
			return nil
		}

		projectRegistry := registry.EnsureProjectExists(message, projectID)
		if projectRegistry != nil {
			projectRegistry.ApplyEventToAllStores(message)
			registry.UpdateEntityLookups(message, projectID)
		}

		return nil
	})))

	return registry
}

func setupLogger(level slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
}
