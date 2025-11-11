package codeview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/lib/test-helpers/reducer-helpers"
	"testing"
)

func createTestCode(projectID string) *Code {
	return &Code{
		ID:        "code-1",
		ProjectID: projectID,
		Slug:      "topic:climate",
		Color:     "green-500",
		Definition: "Climate topics",
	}
}

func TestCodeReducer(t *testing.T) {
	tests := []reducer_helpers.ReducerTestCase[*Code]{
		{
			Name:    "CreatedCode initializes code",
			Initial: nil,
			Event: newCodeEvent("code-1", code.CreatedCode, &code.CreatedCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Definition: "Climate topics",
			}),
			Expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Definition: "Climate topics",
			},
		},
		{
			Name: "UpdatedCode changes color only",
			Initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Definition: "Climate topics",
			},
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color: "emerald-600",
			}),
			Expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "emerald-600",
				Definition: "Climate topics",
			},
		},
		{
			Name: "UpdatedCode changes definition only",
			Initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Definition: "Climate topics",
			},
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Definition: "Renewable energy and sustainability",
			}),
			Expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Definition: "Renewable energy and sustainability",
			},
		},
		{
			Name: "UpdatedCode changes both fields",
			Initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Definition: "Climate topics",
			},
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color:     "teal-500",
				Definition: "Environmental coverage",
			}),
			Expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "teal-500",
				Definition: "Environmental coverage",
			},
		},
		{
			Name: "UpdatedCode with empty strings preserves current values",
			Initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "teal-500",
				Definition: "Environmental coverage",
			},
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color:     "",
				Definition: "",
			}),
			Expected: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "teal-500",
				Definition: "Environmental coverage",
			},
		},
		{
			Name: "MergedCodes returns nil when processing source code",
			Initial: &Code{
				ID:        "code-1",
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Definition: "Climate topics",
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
				Definition: "Temperature topics",
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
				Definition: "Temperature topics",
			},
		},
	}

	deletedEntityTests := reducer_helpers.DeletedEntityTests(
		code.EntityName,
		code.DeletedCode,
		func() *Code { return createTestCode("project-1") },
	)
	deletedProjectTests := reducer_helpers.DeletedProjectCascadeTests(createTestCode)

	combinedTests := append(tests, deletedEntityTests...)
	combinedTests = append(combinedTests, deletedProjectTests...)

	reducer_helpers.RunReducerTests(t, combinedTests, Reducer)
}

func newCodeEvent(aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent(action, payload, code.EntityName, aggregateID, (*commands.AnyMessage)(nil)))
}
