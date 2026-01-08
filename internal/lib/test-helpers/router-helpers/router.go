package router_helpers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/projections/registry"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

type RouterTestCase struct {
	Name            string
	Input           *commands.AnyMessage
	ExpectErr       string
	ExpectEvent     *commands.AnyMessage
	ExpectPublished []*commands.AnyMessage
	IgnoreFields    []th.IgnoreFieldsOption
}

func RunRouterTests(
	t *testing.T,
	setupCommands []*commands.AnyMessage,
	tests []RouterTestCase,
	newRouter func(*registry.Store) dispatch.CommandRouter,
) {
	store := NewTestStore(setupCommands)
	router := newRouter(store)

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
			IgnoreFields:    tt.IgnoreFields,
		}
	}

	dispatch.RunPublisherTests(t, publisherTests)
}
