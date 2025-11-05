package registry_helpers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/registry"
)

// ApplyTestEvents applies multiple events to the registry, simulating event processing
func ApplyTestEvents(reg *registry.ProjectViewRegistry, events []*commands.AnyMessage) {
	for _, event := range events {
		ApplyTestEvent(reg, event)
	}
}

// ApplyTestEvent applies an event to the registry, simulating event processing
func ApplyTestEvent(reg *registry.ProjectViewRegistry, event *commands.AnyMessage) {
	projectID := commands.ExtractProjectID(event)
	if projectID == "" {
		return
	}

	projectView := reg.EnsureProjectExists(event, projectID)
	if projectView != nil {
		projectView.ApplyEventToAllStores(event)
		reg.UpdateEntityLookups(event, projectID)
	}
}
