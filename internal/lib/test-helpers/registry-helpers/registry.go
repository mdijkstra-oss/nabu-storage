package registry_helpers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/registry"
	codeview "hermes-relay/internal/domain/projections/code-entity"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	"testing"
)

type RouterTestCase struct {
	Name            string
	Input           *commands.AnyMessage
	ExpectErr       string
	ExpectEvent     *commands.AnyMessage
	ExpectPublished []*commands.AnyMessage
}

func RunRouterTests(
	t *testing.T,
	setupCommands []*commands.AnyMessage,
	tests []RouterTestCase,
	newRouter func(*registry.ProjectViewRegistry) dispatch.CommandRouter,
) {
	reg := NewTestRegistry(setupCommands)
	router := newRouter(reg)

	publisherTests := make([]dispatch.PublisherTestCase, len(tests))
	for i, tt := range tests {
		var expectedPublished []*commands.AnyMessage
		if tt.ExpectPublished != nil {
			expectedPublished = append([]*commands.AnyMessage{tt.Input}, tt.ExpectPublished...)
		} else if tt.ExpectEvent != nil {
			expectedPublished = []*commands.AnyMessage{tt.Input, tt.ExpectEvent}
		} else {
			expectedPublished = []*commands.AnyMessage{tt.Input}
		}

		publisherTests[i] = dispatch.PublisherTestCase{
			Name:            tt.Name,
			Subscribers:     []dispatch.CommandRouter{router},
			Input:           tt.Input,
			ExpectErr:       tt.ExpectErr,
			ExpectEvent:     tt.ExpectEvent,
			ExpectPublished: expectedPublished,
		}
	}

	dispatch.RunPublisherTests(t, publisherTests)
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
