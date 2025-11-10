package domain_helpers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/test-helpers"
	"testing"
)

// Todo: drop this I think, this function exists in domain code itself
func NewDomainEvent(entityName commands.AggregateType, aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent(action, payload, entityName, aggregateID, (*commands.AnyMessage)(nil)))
}

func AssertMessage(t *testing.T, got, want *commands.AnyMessage, msg string) {
	t.Helper()
	test_helpers.AssertEqualIgnoreFields(t, got, want, msg, commands.AnyMessage{}, "ID", "Timestamp", "CausationID")
}
