package domain

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	documenthandlers "hermes-relay/internal/domain/entities/document/handlers"
	projecthandlers "hermes-relay/internal/domain/entities/project/handlers"
	"hermes-relay/internal/domain/projections/registry"
)

func NewCommandRouter(store *registry.Store) dispatch.CommandRouter {
	return dispatch.LimitOnType(commands.Command,
		registry.EnsureHealth(store,
			dispatch.CombineRouters(
				documenthandlers.NewRouter(store),
				projecthandlers.NewRouter(store),
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

func CommandHandlers(store *registry.Store) []dispatch.CommandRouter {
	return []dispatch.CommandRouter{
		NewCommandRouter(store),
		NewSagaRouter(),
	}
}
