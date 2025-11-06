package registry_helpers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/registry"
	codeview "hermes-relay/internal/domain/projections/code-entity"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

type RouterTestCase struct {
	Name        string
	Input       *commands.AnyMessage
	ExpectErr   string
	ExpectEvent *commands.AnyMessage
}

func RunRouterTests(
	t *testing.T,
	setupCommands []*commands.AnyMessage,
	tests []RouterTestCase,
	newRouter func(*registry.ProjectViewRegistry) dispatch.CommandRouter,
) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			reg := NewTestRegistry(setupCommands)
			result, err := newRouter(reg)(tt.Input, nil)

			th.AssertError(t, err, tt.ExpectErr, "error")
			if tt.ExpectErr == "" {
				th.AssertMessage(t, result, tt.ExpectEvent, "event")
			}
		})
	}
}

func NewTestRegistry(commands []*commands.AnyMessage) *registry.ProjectViewRegistry {
	reg := registry.NewProjectViewRegistry(
		projectview.Reducer,
		codeview.Reducer,
		fileview.Reducer,
	)

	ApplyTestEvents(reg, commands)

	return reg
}

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
