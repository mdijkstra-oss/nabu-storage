package codeview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestCodeReducer(t *testing.T) {
	tests := []th.ReducerTestCase[*Code]{
		{
			Name:    "CreatedCode initializes code",
			Initial: nil,
			Event: newCodeEvent("code-1", code.CreatedCode, &code.CreatedCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			}),
			Expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			},
		},
		{
			Name: "UpdatedCode changes color only",
			Initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			},
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color: "emerald-600",
			}),
			Expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "emerald-600",
				Reasoning: "Climate topics",
			},
		},
		{
			Name: "UpdatedCode changes reasoning only",
			Initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			},
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Reasoning: "Renewable energy and sustainability",
			}),
			Expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Renewable energy and sustainability",
			},
		},
		{
			Name: "UpdatedCode changes both fields",
			Initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			},
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color:     "teal-500",
				Reasoning: "Environmental coverage",
			}),
			Expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "teal-500",
				Reasoning: "Environmental coverage",
			},
		},
		{
			Name: "UpdatedCode with empty strings preserves current values",
			Initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "teal-500",
				Reasoning: "Environmental coverage",
			},
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color:     "",
				Reasoning: "",
			}),
			Expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "teal-500",
				Reasoning: "Environmental coverage",
			},
		},
		{
			Name: "MergedCodes returns nil when processing source code",
			Initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			},
			Event: newCodeEvent("code-1", code.MergedCodes, &code.MergedCodesPayload{
				SourceID: "code-1",
				TargetID: "code-2",
			}),
			Expected: nil,
		},
		{
			Name: "MergedCodes keeps target code unchanged",
			Initial: &Code{
				ID:        "code-2",
				ProjectID: "project-1",
				Slug:      "topic:temperature",
				Color:     "red-500",
				Reasoning: "Temperature topics",
			},
			Event: newCodeEvent("code-2", code.MergedCodes, &code.MergedCodesPayload{
				SourceID: "code-1",
				TargetID: "code-2",
			}),
			Expected: &Code{
				ID:        "code-2",
				ProjectID: "project-1",
				Slug:      "topic:temperature",
				Color:     "red-500",
				Reasoning: "Temperature topics",
			},
		},
		{
			Name: "DeletedCode returns nil",
			Initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Climate topics",
			},
			Event:    newCodeEvent("code-1", code.DeletedCode, nil),
			Expected: nil,
		},
	}

	th.RunReducerTests(t, tests, Reducer)
}

func newCodeEvent(aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent(action, payload, code.EntityName, aggregateID, (*commands.AnyMessage)(nil)))
}
