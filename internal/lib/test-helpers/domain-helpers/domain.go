package domain_helpers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/utils"
)

func TestActor() commands.Actor {
	return commands.Actor{UserID: "test-user", ActorType: commands.ActorTypeSystem}
}

func NewDomainEvent(entityName commands.AggregateType, aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent(action, payload, entityName, aggregateID, TestActor(), (*commands.AnyMessage)(nil)))
}

func NewCommand(entityName commands.AggregateType, aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewCommand(action, payload, entityName, aggregateID, TestActor(), (*commands.AnyMessage)(nil)))
}

func NewCommandWithExpectedVersion(entityName commands.AggregateType, aggregateID string, action commands.Action, payload any, expectedVersion int) *commands.AnyMessage {
	msg := NewCommand(entityName, aggregateID, action, payload)
	msg.ExpectedEntityVersion = &expectedVersion
	return msg
}

func AssertMessage(t *testing.T, got, want *commands.AnyMessage, msg string, extraOpts ...cmp.Option) {
	t.Helper()
	ignoreOpts := []test_helpers.IgnoreFieldsOption{
		{Type: commands.AnyMessage{}, Fields: []string{"ID", "Timestamp", "CausationID", "AggregateID", "Actor"}},
	}
	defaultOpts := test_helpers.ToCmpOptions(ignoreOpts)
	opts := append(defaultOpts, extraOpts...)
	test_helpers.AssertEqual(t, got, want, msg, opts...)
}

// AssertDomainEventMessage verifies domain event has valid AggregateID
// Domain events ALWAYS operate on a specific entity, so AggregateID is required
func AssertDomainEventMessage(t *testing.T, got, want *commands.AnyMessage, msg string, extraOpts ...cmp.Option) {
	t.Helper()

	if got == nil {
		t.Errorf("%s: got nil message", msg)
		return
	}

	if got.Type != commands.DomainEvent {
		t.Errorf("%s: expected DomainEvent, got %s", msg, got.Type)
	}

	if !utils.ValidAggregateID(string(got.AggregateType), got.AggregateID) {
		t.Errorf("%s: AggregateID must be valid for %s, got %s", msg, got.AggregateType, got.AggregateID)
	}

	AssertMessage(t, got, want, msg, extraOpts...)
}
