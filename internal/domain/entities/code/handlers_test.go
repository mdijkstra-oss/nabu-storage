package code

import (
	"hermes-relay/internal/cqrs"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestCodeRouter(t *testing.T) {
	// Specific test cases
	tests := []th.RouterTestCase{
		th.CommandToEventCase("CreateCode with valid payload", CreateCode, CreatedCode, CreateCodePayload{
			Slug:      "topic:climate",
			Color:     "green-500",
			Reasoning: "Environmental topics",
		}, EntityName, ""),

		th.CommandToEventCase("UpdateCode with valid payload", UpdateCode, UpdatedCode, UpdateCodePayload{
			Color:     "emerald-600",
			Reasoning: "Updated environmental coverage",
		}, EntityName, "code-123"),

		th.CommandToEventCase("UpdateCode with partial payload (color only)", UpdateCode, UpdatedCode, UpdateCodePayload{
			Color: "teal-500",
		}, EntityName, "code-456"),

		th.CommandToEventCase("UpdateCode with partial payload (reasoning only)", UpdateCode, UpdatedCode, UpdateCodePayload{
			Reasoning: "Comprehensive climate coverage",
		}, EntityName, "code-789"),

		th.CommandToEventCase[any]("DeleteCode", DeleteCode, DeletedCode, nil, EntityName, "code-999"),

		{
			Name: "CreateCode with missing required fields",
			InputMessage: cqrs.ToAny(cqrs.NewCommand[CreateCodePayload, any](CreateCode, CreateCodePayload{
				Slug: "topic:incomplete",
				// Missing Color and Reasoning
			}, EntityName, "", nil)),
			ExpectedReturn:    nil,
			ExpectedPublished: nil,
			ExpectError:       true,
		},
	}

	// Add common router test cases (wrong entity, wrong message type, wrong action)
	tests = append(tests, th.CommonRouterTestCases(CreateCode, CreateCodePayload{
		Slug:      "topic:test",
		Color:     "blue-500",
		Reasoning: "Test",
	}, EntityName)...)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			th.TestRouter(t, Router, tt)
		})
	}
}
