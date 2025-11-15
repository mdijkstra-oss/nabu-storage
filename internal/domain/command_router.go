package domain

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	codehandlers "hermes-relay/internal/domain/entities/code/handlers"
	filehandlers "hermes-relay/internal/domain/entities/file/handlers"
	projecthandlers "hermes-relay/internal/domain/entities/project/handlers"
	"hermes-relay/internal/domain/projections/registry"
)

func NewCommandRouter(registryState *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.LimitOnType(commands.Command,
		registry.EnsureProjectHealth(registryState,
			registry.EnsureEntityHealth(registryState,
				dispatch.CombineRouters(
					codehandlers.NewRouter(registryState),
					filehandlers.NewRouter(registryState),
					projecthandlers.NewRouter(registryState),
				),
			),
		),
	)
}
