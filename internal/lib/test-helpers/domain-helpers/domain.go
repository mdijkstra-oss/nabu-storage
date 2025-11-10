package domain_helpers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/utils"
	"testing"
)

// Todo: drop this I think, this function exists in domain code itself
func NewDomainEvent(entityName commands.AggregateType, aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent(action, payload, entityName, aggregateID, (*commands.AnyMessage)(nil)))
}

func AssertMessage(t *testing.T, got, want *commands.AnyMessage, msg string) {
	t.Helper()
	test_helpers.AssertEqualIgnoreFields(t, got, want, msg, commands.AnyMessage{}, "ID", "Timestamp", "CausationID", "AggregateID")
}

// AssertDomainEventMessage verifies domain event has valid AggregateID
// Domain events ALWAYS operate on a specific entity, so AggregateID is required
func AssertDomainEventMessage(t *testing.T, got, want *commands.AnyMessage, msg string) {
	t.Helper()

	if got == nil {
		t.Errorf("%s: got nil message", msg)
		return
	}

	if got.Type != commands.DomainEvent {
		t.Errorf("%s: expected DomainEvent, got %s", msg, got.Type)
	}

	if !utils.ValidID(got.AggregateID) {
		t.Errorf("%s: AggregateID must be valid UUID, got %s", msg, got.AggregateID)
	}

	AssertMessage(t, got, want, msg)
}
