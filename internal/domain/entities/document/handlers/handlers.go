package handlers

import (
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/projections/registry"
)

func NewRouter(store *registry.Store) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		NewDocumentRouter(store),
		NewBlockRouter(store),
	)
}
