package codeview

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/code"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

// newCodeDomainEvent creates a domain event for code entity testing
func newCodeDomainEvent(aggregateID string, action cqrs.Action, payload any) *cqrs.AnyMessage {
	return cqrs.ToAny(cqrs.NewDomainEvent(action, payload, code.EntityName, aggregateID, (*cqrs.AnyMessage)(nil)))
}

func TestCodeReducer_FullLifecycle(t *testing.T) {
	aggregateID := "test-aggregate-123"
	var state *Code

	// Step 1: Create
	state = Reducer(state, newCodeDomainEvent(aggregateID, code.CreatedCode, &code.CreatedCodePayload{
		Slug:      "topic:climate-change",
		Color:     "green-500",
		Reasoning: "Initial climate topics",
	}))

	th.AssertEqual(t, state, &Code{
		ID:        aggregateID,
		Slug:      "topic:climate-change",
		Color:     "green-500",
		Reasoning: "Initial climate topics",
	}, "After create")

	// Step 2: Update color only
	state = Reducer(state, newCodeDomainEvent(aggregateID, code.UpdatedCode, &code.UpdatedCodePayload{
		Color: "emerald-600",
	}))

	th.AssertEqual(t, state, &Code{
		ID:        aggregateID,
		Slug:      "topic:climate-change",
		Color:     "emerald-600",
		Reasoning: "Initial climate topics",
	}, "After update color")

	// Step 3: Update reasoning only
	state = Reducer(state, newCodeDomainEvent(aggregateID, code.UpdatedCode, &code.UpdatedCodePayload{
		Reasoning: "Expanded to renewable energy and sustainability",
	}))

	th.AssertEqual(t, state, &Code{
		ID:        aggregateID,
		Slug:      "topic:climate-change",
		Color:     "emerald-600",
		Reasoning: "Expanded to renewable energy and sustainability",
	}, "After update reasoning")

	// Step 4: Update both fields
	state = Reducer(state, newCodeDomainEvent(aggregateID, code.UpdatedCode, &code.UpdatedCodePayload{
		Color:     "teal-500",
		Reasoning: "Final comprehensive environmental coverage",
	}))

	th.AssertEqual(t, state, &Code{
		ID:        aggregateID,
		Slug:      "topic:climate-change",
		Color:     "teal-500",
		Reasoning: "Final comprehensive environmental coverage",
	}, "After update both")

	// Step 5: Update with empty strings (should not change anything)
	state = Reducer(state, newCodeDomainEvent(aggregateID, code.UpdatedCode, &code.UpdatedCodePayload{
		Color:     "",
		Reasoning: "",
	}))

	th.AssertEqual(t, state, &Code{
		ID:        aggregateID,
		Slug:      "topic:climate-change",
		Color:     "teal-500",
		Reasoning: "Final comprehensive environmental coverage",
	}, "After empty update")

	// Step 6: Delete
	state = Reducer(state, newCodeDomainEvent(aggregateID, code.DeletedCode, nil))

	th.AssertEqual(t, state, nil, "After delete")
}
