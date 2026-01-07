package domain

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	documenthandlers "hermes-relay/internal/domain/entities/document/handlers"
	projecthandlers "hermes-relay/internal/domain/entities/project/handlers"
	"hermes-relay/internal/domain/projections/registry"
)

func NewCommandRouter(registryState *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.LimitOnType(commands.Command,
		registry.EnsureHealth(registryState,
			dispatch.CombineRouters(
				documenthandlers.NewRouter(registryState),
				projecthandlers.NewRouter(registryState),
			),
		),
	)
}

func NewSagaRouter() dispatch.CommandRouter {
	return dispatch.LimitOnType(commands.DomainEvent,
		dispatch.CombineRouters(
			projecthandlers.DefaultDocumentsSaga(),
		),
	)
}

func CommandHandlers(registryState *registry.RegistryState) []dispatch.CommandRouter {
	return []dispatch.CommandRouter{
		NewCommandRouter(registryState),
		NewSagaRouter(),
	}
}
