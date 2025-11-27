package dispatch_test

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/projections/registry"
	router_helpers "hermes-relay/internal/lib/test-helpers/router-helpers"
	"hermes-relay/internal/lib/utils"
	"testing"
)

const (
	TestCommand1 commands.Action = "TestCommand1"
	TestCommand2 commands.Action = "TestCommand2"
	TestCommand3 commands.Action = "TestCommand3"
	TestEvent1   commands.Action = "TestEvent1"
	TestEvent2   commands.Action = "TestEvent2"
	TestEvent3   commands.Action = "TestEvent3"
	TestEvent4   commands.Action = "TestEvent4"
)

func testActorRouter(_ *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		dispatch.LimitOnAction(TestCommand1, func(msg *commands.AnyMessage, _ dispatch.PublishFunc) (*commands.AnyMessage, error) {
			return commands.ToDomainEvent(msg, TestEvent1), nil
		}),
		dispatch.LimitOnAction(TestCommand2, func(msg *commands.AnyMessage, publish dispatch.PublishFunc) (*commands.AnyMessage, error) {
			event2 := commands.ToDomainEvent(msg, TestEvent2)
			event3 := commands.ToDomainEvent(event2, TestEvent3)
			_, _ = publish(event3)
			return event2, nil
		}),
		dispatch.LimitOnAction(TestCommand3, func(msg *commands.AnyMessage, _ dispatch.PublishFunc) (*commands.AnyMessage, error) {
			return commands.ToDomainEvent(msg, TestEvent4), nil
		}),
	)
}

func TestActorPropagationThroughSystem(t *testing.T) {
	humanActor := commands.Actor{UserID: "user-123", ActorType: commands.ActorTypeHuman}
	llmActor := commands.Actor{UserID: "claude", ActorType: commands.ActorTypeLLM}
	systemActor := commands.Actor{UserID: "system", ActorType: commands.ActorTypeSystem}

	testID1 := utils.NewID()
	testID2 := utils.NewID()
	testID3 := utils.NewID()
	causeID := utils.NewID()

	causeEvent := eventWithActor(TestEvent1, causeID, systemActor, nil)

	router_helpers.RunRouterTests(t,
		nil,
		[]router_helpers.RouterTestCase{
			{
				Name:        "Returned event preserves human actor",
				Input:       commandWithActor(TestCommand1, testID1, humanActor, nil),
				ExpectEvent: eventWithActor(TestEvent1, testID1, humanActor, nil),
			},
			{
				Name:        "Dispatched event preserves LLM actor",
				Input:       commandWithActor(TestCommand2, testID2, llmActor, nil),
				ExpectEvent: eventWithActor(TestEvent2, testID2, llmActor, nil),
				ExpectPublished: []*commands.AnyMessage{
					eventWithActor(TestEvent3, testID2, llmActor, nil),
					eventWithActor(TestEvent2, testID2, llmActor, nil),
				},
			},
			{
				Name:        "Causation chain preserves system actor",
				Input:       commandWithActor(TestCommand3, testID3, systemActor, causeEvent),
				ExpectEvent: eventWithActor(TestEvent4, testID3, systemActor, causeEvent),
			},
		},
		testActorRouter,
	)
}

func commandWithActor(action commands.Action, aggregateID string, actor commands.Actor, cause *commands.AnyMessage) *commands.AnyMessage {
	return commands.ToAny(commands.NewCommand[any, any](
		action,
		nil,
		"TestAggregate",
		aggregateID,
		actor,
		cause,
	))
}

func eventWithActor(action commands.Action, aggregateID string, actor commands.Actor, cause *commands.AnyMessage) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent[any, any](
		action,
		nil,
		"TestAggregate",
		aggregateID,
		actor,
		cause,
	))
}
