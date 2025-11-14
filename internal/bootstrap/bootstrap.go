package bootstrap

import (
	"fmt"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	codehandlers "hermes-relay/internal/domain/entities/code/handlers"
	filehandlers "hermes-relay/internal/domain/entities/file/handlers"
	projecthandlers "hermes-relay/internal/domain/entities/project/handlers"
	"hermes-relay/internal/domain/projections/registry"
	"log/slog"
	"os"
)

func SetupLogger(level slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
}

func SetupRegistry(publisher *dispatch.InMemoryPublisher) *registry.RegistryState {
	registryState := registry.NewRegistryState()

	publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, dispatch.ReadOnlyRoutes(func(message *commands.AnyMessage) error {
		projectID := commands.ExtractProjectID(message)
		if projectID == "" {
			projectID = registryState.GetProjectIDForEntity(message.AggregateType, message.AggregateID)
		}

		if projectID == "" {
			return fmt.Errorf("required project ID for any domain event. %+v", message)
		}

		registryState.ApplyEvent(message)

		return nil
	})))

	return registryState
}

func SetupCommandHandlers(publisher *dispatch.InMemoryPublisher, registryState *registry.RegistryState) {
	slog.Info("Setting up command handlers for new incoming messages")

	var commandRouter = dispatch.LimitOnType(commands.Command,
		dispatch.CombineRouters(
			codehandlers.NewRouter(registryState),
			filehandlers.NewRouter(registryState),
			projecthandlers.NewRouter(registryState),
		),
	)

	publisher.Subscribe(commandRouter)
}
