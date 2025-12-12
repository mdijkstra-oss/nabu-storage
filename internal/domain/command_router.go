package domain

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	charthandlers "hermes-relay/internal/domain/entities/chart/handlers"
	codehandlers "hermes-relay/internal/domain/entities/code/handlers"
	filehandlers "hermes-relay/internal/domain/entities/file/handlers"
	projecthandlers "hermes-relay/internal/domain/entities/project/handlers"
	"hermes-relay/internal/domain/projections/registry"
)

func NewCommandRouter(registryState *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.LimitOnType(commands.Command,
		registry.EnsureProjectHealth(registryState,
			registry.EnsureEntityHealth(registryState,
				//registry.EnsureExpectedVersion(registryState, // temp disable due to token increase
				// eg if user changes 1 letter, llm does large operation, bam rejected
				// so prefer last write then
				// perhaps need to check that user may delete thing before llm sends batch todo: that
				dispatch.CombineRouters(
					charthandlers.NewRouter(registryState),
					codehandlers.NewRouter(registryState),
					filehandlers.NewRouter(registryState),
					projecthandlers.NewRouter(registryState),
				),
				//),
			),
		),
	)
}

func NewSagaRouter() dispatch.CommandRouter {
	return dispatch.LimitOnType(commands.DomainEvent,
		dispatch.CombineRouters(
			projecthandlers.DefaultFilesSaga(),
		),
	)
}

func CommandHandlers(registryState *registry.RegistryState) []dispatch.CommandRouter {
	return []dispatch.CommandRouter{
		NewCommandRouter(registryState),
		NewSagaRouter(),
	}
}
