package test_helpers

import (
	"hermes-relay/internal/cqrs"
	"testing"
)

type RouterTestCase struct {
	Name              string
	InputMessage      *cqrs.AnyMessage
	ExpectedReturn    *cqrs.AnyMessage
	ExpectedPublished []*cqrs.AnyMessage
	ExpectError       bool
}

func TestRouter(t *testing.T, router cqrs.CommandRouter, tc RouterTestCase) {
	t.Helper()

	var published []*cqrs.AnyMessage
	mockPublisher := func(event *cqrs.AnyMessage) (*cqrs.AnyMessage, error) {
		published = append(published, event)
		return event, nil
	}

	result, err := router(tc.InputMessage, mockPublisher)

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

func CommandToEventCase[T any](name string, commandAction, eventAction cqrs.Action, payload T, entityType cqrs.AggregateType, aggregateID string) RouterTestCase {
	return RouterTestCase{
		Name:              name,
		InputMessage:      cqrs.ToAny(cqrs.NewCommand[T, any](commandAction, payload, entityType, aggregateID, nil)),
		ExpectedReturn:    cqrs.ToAny(cqrs.NewDomainEvent[T, any](eventAction, payload, entityType, aggregateID, nil)),
		ExpectedPublished: nil,
	}
}

func WrongEntityTypeCase[T any](action cqrs.Action, payload T, correctEntityType cqrs.AggregateType) RouterTestCase {
	return RouterTestCase{
		Name:              "Wrong entity type should return nil",
		InputMessage:      cqrs.ToAny(cqrs.NewCommand[T, any](action, payload, "DifferentEntity", "", nil)),
		ExpectedReturn:    nil,
		ExpectedPublished: nil,
	}
}

func WrongMessageTypeCase[T any](action cqrs.Action, payload T, entityType cqrs.AggregateType, aggregateID string) RouterTestCase {
	return RouterTestCase{
		Name:              "Wrong message type should return nil",
		InputMessage:      cqrs.ToAny(cqrs.NewDomainEvent[T, any](action, payload, entityType, aggregateID, nil)),
		ExpectedReturn:    nil,
		ExpectedPublished: nil,
	}
}

func WrongActionCase[T any](correctAction cqrs.Action, payload T, entityType cqrs.AggregateType, aggregateID string) RouterTestCase {
	return RouterTestCase{
		Name:              "Wrong action should return nil",
		InputMessage:      cqrs.ToAny(cqrs.NewCommand[T, any]("DifferentAction", payload, entityType, aggregateID, nil)),
		ExpectedReturn:    nil,
		ExpectedPublished: nil,
	}
}

func ValidationErrorCase[T any](name string, action cqrs.Action, payload T, entityType cqrs.AggregateType, aggregateID string) RouterTestCase {
	return RouterTestCase{
		Name:              name,
		InputMessage:      cqrs.ToAny(cqrs.NewCommand[T, any](action, payload, entityType, aggregateID, nil)),
		ExpectedReturn:    nil,
		ExpectedPublished: nil,
		ExpectError:       true,
	}
}
