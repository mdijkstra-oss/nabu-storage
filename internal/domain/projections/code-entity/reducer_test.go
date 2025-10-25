package codeview

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/lib/test-helpers"
	"testing"
)

// NewCodeEvent creates a domain event for code entity testing
func NewCodeEvent(aggregateID string, action cqrs.Action, payload any) *cqrs.AnyMessage {
	return cqrs.ToAny(cqrs.NewDomainEvent(action, payload, code.EntityName, aggregateID, (*cqrs.AnyMessage)(nil)))
}

func TestCodeReducer_FullLifecycle(t *testing.T) {
	aggregateID := "test-aggregate-123"
	var state *Code

	// Step 1: Create
	state = Reducer(state, NewCodeEvent(aggregateID, code.CreatedCode, &code.CreatedCodePayload{
		Slug:      "topic:climate-change",
		Color:     "green-500",
		Reasoning: "Initial climate topics",
	}))

	th.AssertStructEquality(t, state, &Code{
		ID:        aggregateID,
		Slug:      "topic:climate-change",
		Color:     "green-500",
		Reasoning: "Initial climate topics",
	}, "After create")

	// Step 2: Update color only
	state = Reducer(state, NewCodeEvent(aggregateID, code.UpdatedCode, &code.UpdatedCodePayload{
		Color: "emerald-600",
	}))

	th.AssertStructEquality(t, state, &Code{
		ID:        aggregateID,
		Slug:      "topic:climate-change",
		Color:     "emerald-600",
		Reasoning: "Initial climate topics",
	}, "After update color")

	// Step 3: Update reasoning only
	state = Reducer(state, NewCodeEvent(aggregateID, code.UpdatedCode, &code.UpdatedCodePayload{
		Reasoning: "Expanded to renewable energy and sustainability",
	}))

	th.AssertStructEquality(t, state, &Code{
		ID:        aggregateID,
		Slug:      "topic:climate-change",
		Color:     "emerald-600",
		Reasoning: "Expanded to renewable energy and sustainability",
	}, "After update reasoning")

	// Step 4: Update both fields
	state = Reducer(state, NewCodeEvent(aggregateID, code.UpdatedCode, &code.UpdatedCodePayload{
		Color:     "teal-500",
		Reasoning: "Final comprehensive environmental coverage",
	}))

	th.AssertStructEquality(t, state, &Code{
		ID:        aggregateID,
		Slug:      "topic:climate-change",
		Color:     "teal-500",
		Reasoning: "Final comprehensive environmental coverage",
	}, "After update both")

	// Step 5: Update with empty strings (should not change anything)
	state = Reducer(state, NewCodeEvent(aggregateID, code.UpdatedCode, &code.UpdatedCodePayload{
		Color:     "",
		Reasoning: "",
	}))

	th.AssertStructEquality(t, state, &Code{
		ID:        aggregateID,
		Slug:      "topic:climate-change",
		Color:     "teal-500",
		Reasoning: "Final comprehensive environmental coverage",
	}, "After empty update")

	// Step 6: Delete
	state = Reducer(state, NewCodeEvent(aggregateID, code.DeletedCode, nil))

	th.AssertStructEquality(t, state, nil, "After delete")
}
