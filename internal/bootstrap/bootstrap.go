package bootstrap

import (
	"fmt"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/patches"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain"
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

func SetupRegistryForReplay(publisher *dispatch.InMemoryPublisher) (*registry.Store, func()) {
	store := registry.NewStore()
	unsubscribe := subscribeReplayApplier(publisher, store)
	return store, unsubscribe
}

func SetupRegistryPatching(publisher *dispatch.InMemoryPublisher, store *registry.Store, activeChecker patches.ActiveProjectChecker) {
	subscribePatchEmitter(publisher, store, activeChecker)
}

func subscribeReplayApplier(publisher *dispatch.InMemoryPublisher, store *registry.Store) func() {
	return publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, func(message *commands.AnyMessage, _ dispatch.PublishFunc) (*commands.AnyMessage, error) {
		projectID := projection.Read(store, func(r *registry.Registry) string {
			return registry.ResolveProjectID(r, message)
		})
		if projectID == "" {
			return nil, fmt.Errorf("required project ID for any domain event. %+v", message)
		}
		projection.Apply(store, message)
		return nil, nil
	}))
}

func subscribePatchEmitter(
	publisher *dispatch.InMemoryPublisher,
	store *registry.Store,
	activeChecker patches.ActiveProjectChecker,
) {
	publisher.Subscribe(dispatch.LimitOnType(commands.DomainEvent, func(message *commands.AnyMessage, pub dispatch.PublishFunc) (*commands.AnyMessage, error) {
		projectID := projection.Read(store, func(r *registry.Registry) string {
			return registry.ResolveProjectID(r, message)
		})
		if projectID == "" {
			return nil, fmt.Errorf("required project ID for any domain event. %+v", message)
		}

		before := projection.Read(store, func(r *registry.Registry) *registry.Registry {
			return r
		})
		beforeProj := registry.GetProject(before, projectID)

		projection.Apply(store, message)

		after := projection.Read(store, func(r *registry.Registry) *registry.Registry {
			return r
		})
		afterProj := registry.GetProject(after, projectID)

		action, err := patches.DecidePatch(beforeProj, afterProj, activeChecker.IsActive(projectID))
		if err != nil {
			slog.Error("failed to decide patch action", "projectID", projectID, "error", err)
		} else {
			patches.EmitPatchAction(pub, projectID, action)
		}

		return nil, nil
	}))
}

func SetupCommandHandlers(publisher *dispatch.InMemoryPublisher, store *registry.Store) {
	slog.Info("Setting up command handlers for new incoming messages")
	for _, router := range domain.CommandHandlers(store) {
		publisher.Subscribe(router)
	}
}
