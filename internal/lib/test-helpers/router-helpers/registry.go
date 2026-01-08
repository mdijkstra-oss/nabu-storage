package router_helpers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/projections/registry"
)

func NewTestStore(events []*commands.AnyMessage) *registry.Store {
	store := registry.NewStore()
	ApplyTestEvents(store, events)
	return store
}

func ApplyTestEvents(store *registry.Store, events []*commands.AnyMessage) {
	for _, event := range events {
		ApplyTestEvent(store, event)
	}
}

func ApplyTestEvent(store *registry.Store, event *commands.AnyMessage) {
	projection.Apply(store, event)
}
