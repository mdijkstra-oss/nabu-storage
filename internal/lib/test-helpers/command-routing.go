package test_helpers

import (
	"context"
	"hermes-relay/internal/cqrs"
	"testing"
)

// RouterTestCase defines a single test case for a CommandRouter
type RouterTestCase struct {
	Name              string
	InputMessage      *cqrs.AnyMessage
	ExpectedReturn    *cqrs.AnyMessage
	ExpectedPublished []*cqrs.AnyMessage
	ExpectError       bool
}

// TestRouter executes a CommandRouter test case
func TestRouter(t *testing.T, router cqrs.CommandRouter, tc RouterTestCase) {
	t.Helper()

	var published []*cqrs.AnyMessage
	mockPublisher := func(ctx context.Context, event *cqrs.AnyMessage) (*cqrs.AnyMessage, error) {
		published = append(published, event)
		return event, nil
	}

	result, err := router(context.Background(), tc.InputMessage, mockPublisher)

	if tc.ExpectError {
		if err == nil {
			t.Fatalf("%s: expected error but got nil", tc.Name)
		}
		return
	}

	if err != nil {
		t.Fatalf("%s: unexpected error: %v", tc.Name, err)
	}

	AssertMessage(t, result, tc.ExpectedReturn, tc.Name+" return")

	if len(published) != len(tc.ExpectedPublished) {
		t.Fatalf("%s: expected %d published events, got %d", tc.Name, len(tc.ExpectedPublished), len(published))
	}

	for i, expected := range tc.ExpectedPublished {
		AssertMessage(t, published[i], expected, tc.Name+" published["+string(rune(i))+"]")
	}
}

// WrongEntityTypeCase creates a test case verifying a router ignores commands for different entities
func WrongEntityTypeCase[T any](action cqrs.Action, payload T, correctEntityType cqrs.AggregateType) RouterTestCase {
	return RouterTestCase{
		Name:              "Wrong entity type should return nil",
		InputMessage:      cqrs.ToAny(cqrs.NewCommand[T, any](action, payload, "DifferentEntity", "", nil)),
		ExpectedReturn:    nil,
		ExpectedPublished: nil,
	}
}

// WrongMessageTypeCase creates a test case verifying a router ignores non-command messages
func WrongMessageTypeCase[T any](action cqrs.Action, payload T, entityType cqrs.AggregateType, aggregateID string) RouterTestCase {
	return RouterTestCase{
		Name:              "Wrong message type should return nil",
		InputMessage:      cqrs.ToAny(cqrs.NewDomainEvent[T, any](action, payload, entityType, aggregateID, nil)),
		ExpectedReturn:    nil,
		ExpectedPublished: nil,
	}
}

// WrongActionCase creates a test case verifying a router ignores commands with different actions
func WrongActionCase[T any](correctAction cqrs.Action, payload T, entityType cqrs.AggregateType, aggregateID string) RouterTestCase {
	return RouterTestCase{
		Name:              "Wrong action should return nil",
		InputMessage:      cqrs.ToAny(cqrs.NewCommand[T, any]("DifferentAction", payload, entityType, aggregateID, nil)),
		ExpectedReturn:    nil,
		ExpectedPublished: nil,
	}
}

// CommonRouterTestCases generates standard test cases for wrong entity, wrong message type, and wrong action
func CommonRouterTestCases[T any](action cqrs.Action, payload T, entityType cqrs.AggregateType) []RouterTestCase {
	return []RouterTestCase{
		WrongEntityTypeCase(action, payload, entityType),
		WrongMessageTypeCase(action, payload, entityType, "test-aggregate-id"),
		WrongActionCase(action, payload, entityType, "test-aggregate-id"),
	}
}
