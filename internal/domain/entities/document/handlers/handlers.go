package handlers

import (
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/projections/registry"
)

func NewRouter(registryState *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		NewDocumentRouter(registryState),
		NewBlockRouter(registryState),
	)
}
