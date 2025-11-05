package codeview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestCodeReducer(t *testing.T) {
	tests := []struct {
		name     string
		initial  *Code
		event    *commands.AnyMessage
		expected *Code
	}{
		{
			name:    "CreatedCode initializes code",
			initial: nil,
			event: newCodeEvent("code-1", code.CreatedCode, &code.CreatedCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			}),
			expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			},
		},
		{
			name: "UpdatedCode changes color only",
			initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			},
			event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color: "emerald-600",
			}),
			expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "emerald-600",
				Reasoning: "Climate topics",
			},
		},
		{
			name: "UpdatedCode changes reasoning only",
			initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			},
			event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Reasoning: "Renewable energy and sustainability",
			}),
			expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Renewable energy and sustainability",
			},
		},
		{
			name: "UpdatedCode changes both fields",
			initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			},
			event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color:     "teal-500",
				Reasoning: "Environmental coverage",
			}),
			expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "teal-500",
				Reasoning: "Environmental coverage",
			},
		},
		{
			name: "UpdatedCode with empty strings preserves current values",
			initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "teal-500",
				Reasoning: "Environmental coverage",
			},
			event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color:     "",
				Reasoning: "",
			}),
			expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "teal-500",
				Reasoning: "Environmental coverage",
			},
		},
		{
			name: "DeletedCode returns nil",
			initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			},
			event:    newCodeEvent("code-1", code.DeletedCode, nil),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reducer(tt.initial, tt.event)
			th.AssertEqual(t, result, tt.expected, "state after reduction")
		})
	}
}

func newCodeEvent(aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent(action, payload, code.EntityName, aggregateID, (*commands.AnyMessage)(nil)))
}
