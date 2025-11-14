package bootstrap

import (
	"fmt"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/patches"
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

func SetupRegistry(publisher *dispatch.InMemoryPublisher, activeChecker patches.ActiveProjectChecker) *registry.RegistryState {
	registryState := registry.NewRegistryState()

	setupRegistryWithPatching(publisher, registryState, activeChecker)

	return registryState
}

// Todo: may need improvement
// Registry hold project aggregated references
// Patching are events websocket can listen on to get updated and full state
func setupRegistryWithPatching(
	publisher *dispatch.InMemoryPublisher,
	registryState *registry.RegistryState,
	activeChecker patches.ActiveProjectChecker,
) {
	publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, func(message *commands.AnyMessage, pub dispatch.PublishFunc) (*commands.AnyMessage, error) {
		projectID := registryState.ResolveProjectID(message)
		if projectID == "" {
			return nil, fmt.Errorf("required project ID for any domain event. %+v", message)
		}

		before := registryState.GetProject(projectID)
		registryState.ApplyEvent(message)
		after := registryState.GetProject(projectID)

		action, err := patches.DecidePatch(before, after, activeChecker.IsActive(projectID))
		if err != nil {
			slog.Error("failed to decide patch action", "projectID", projectID, "error", err)
		} else {
			patches.EmitPatchAction(pub, projectID, action)
		}

		return nil, nil
	}))
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
