package router_helpers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/projections/registry"
)

func NewTestRegistry(events []*commands.AnyMessage) *registry.RegistryState {
	reg := registry.NewRegistryState()
	ApplyTestEvents(reg, events)
	return reg
}

func ApplyTestEvents(reg *registry.RegistryState, events []*commands.AnyMessage) {
	for _, event := range events {
		ApplyTestEvent(reg, event)
	}
}

func ApplyTestEvent(reg *registry.RegistryState, event *commands.AnyMessage) {
	reg.ApplyEvent(event)
}
