package code

import (
	"hermes-relay/internal/cqrs"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestCodeRouter(t *testing.T) {
	// Specific test cases
	tests := []th.RouterTestCase{
		{
			Name: "CreateCode with valid payload",
			InputMessage: cqrs.ToAny(cqrs.NewCommand[CreateCodePayload, any](CreateCode, CreateCodePayload{
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Environmental topics",
			}, EntityName, "", nil)),
			ExpectedReturn: cqrs.ToAny(cqrs.NewDomainEvent[CreateCodePayload, any](CreatedCode, CreateCodePayload{
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Environmental topics",
			}, EntityName, "", nil)),
			ExpectedPublished: nil,
		},
		{
			Name: "UpdateCode with valid payload",
			InputMessage: cqrs.ToAny(cqrs.NewCommand[UpdateCodePayload, any](UpdateCode, UpdateCodePayload{
				Color:     "emerald-600",
				Reasoning: "Updated environmental coverage",
			}, EntityName, "code-123", nil)),
			ExpectedReturn: cqrs.ToAny(cqrs.NewDomainEvent[UpdateCodePayload, any](UpdatedCode, UpdateCodePayload{
				Color:     "emerald-600",
				Reasoning: "Updated environmental coverage",
			}, EntityName, "code-123", nil)),
			ExpectedPublished: nil,
		},
		{
			Name: "UpdateCode with partial payload (color only)",
			InputMessage: cqrs.ToAny(cqrs.NewCommand[UpdateCodePayload, any](UpdateCode, UpdateCodePayload{
				Color: "teal-500",
			}, EntityName, "code-456", nil)),
			ExpectedReturn: cqrs.ToAny(cqrs.NewDomainEvent[UpdateCodePayload, any](UpdatedCode, UpdateCodePayload{
				Color: "teal-500",
			}, EntityName, "code-456", nil)),
			ExpectedPublished: nil,
		},
		{
			Name: "UpdateCode with partial payload (reasoning only)",
			InputMessage: cqrs.ToAny(cqrs.NewCommand[UpdateCodePayload, any](UpdateCode, UpdateCodePayload{
				Reasoning: "Comprehensive climate coverage",
			}, EntityName, "code-789", nil)),
			ExpectedReturn: cqrs.ToAny(cqrs.NewDomainEvent[UpdateCodePayload, any](UpdatedCode, UpdateCodePayload{
				Reasoning: "Comprehensive climate coverage",
			}, EntityName, "code-789", nil)),
			ExpectedPublished: nil,
		},
		{
			Name:           "DeleteCode",
			InputMessage:   cqrs.ToAny(cqrs.NewCommand[any, any](DeleteCode, nil, EntityName, "code-999", nil)),
			ExpectedReturn: cqrs.ToAny(cqrs.NewDomainEvent[any, any](DeletedCode, nil, EntityName, "code-999", nil)),
			ExpectedPublished: nil,
		},
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
