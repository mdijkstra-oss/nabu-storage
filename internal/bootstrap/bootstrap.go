package bootstrap

import (
	"fmt"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	domainprojection "hermes-relay/internal/cqrs/registry"
	codehandlers "hermes-relay/internal/domain/entities/code/handlers"
	filehandlers "hermes-relay/internal/domain/entities/file/handlers"
	projecthandlers "hermes-relay/internal/domain/entities/project/handlers"
	codeview "hermes-relay/internal/domain/projections/code-entity"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	"log/slog"
	"os"
)

func SetupLogger(level slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
}

func SetupProjectViewRegistry(publisher *dispatch.InMemoryPublisher) *domainprojection.ProjectViewRegistry {
	registry := domainprojection.NewProjectViewRegistry(
		projectview.Reducer,
		codeview.Reducer,
		fileview.Reducer,
	)

	publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, dispatch.ReadOnlyRoutes(func(message *commands.AnyMessage) error {

		projectID := commands.ExtractProjectID(message)
		if projectID == "" {
			projectID = registry.GetProjectIDForEntity(message.AggregateType, message.AggregateID)
		}

		if projectID == "" {
			return fmt.Errorf("required project ID for any domain event. %+v", message)
		}

		projectRegistry := registry.EnsureProjectExists(message, projectID)
		if projectRegistry != nil {
			projectRegistry.ApplyEventToAllStores(message)
			registry.UpdateEntityLookups(message, projectID)
		}

		if message.Action == "DeletedProject" && message.AggregateType == "Project" {
			registry.RemoveProject(projectID)
		}

		return nil
	})))

	return registry
}

func SetupCommandHandlers(publisher *dispatch.InMemoryPublisher, registry *domainprojection.ProjectViewRegistry) {
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
