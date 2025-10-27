package code

import (
	"hermes-relay/internal/cqrs"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestCodeRouter(t *testing.T) {
	tests := []struct {
		name        string
		input       *cqrs.AnyMessage
		expectErr   bool
		expectEvent *cqrs.AnyMessage
	}{
		{
			name: "CreateCode with valid payload",
			input: cqrs.ToAny(cqrs.NewCommand[CreateCodePayload, any](CreateCode, CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Environmental topics",
			}, EntityName, "", nil)),
			expectErr: false,
			expectEvent: cqrs.ToAny(cqrs.NewDomainEvent[CreatedCodePayload, any](CreatedCode, CreatedCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Reasoning: "Environmental topics",
			}, EntityName, "", nil)),
		},
		{
			name: "UpdateCode with both fields",
			input: cqrs.ToAny(cqrs.NewCommand[UpdateCodePayload, any](UpdateCode, UpdateCodePayload{
				Color:     "emerald-600",
				Reasoning: "Updated environmental coverage",
			}, EntityName, "code-123", nil)),
			expectErr: false,
			expectEvent: cqrs.ToAny(cqrs.NewDomainEvent[UpdatedCodePayload, any](UpdatedCode, UpdatedCodePayload{
				Color:     "emerald-600",
				Reasoning: "Updated environmental coverage",
			}, EntityName, "code-123", nil)),
		},
		{
			name: "UpdateCode with color only",
			input: cqrs.ToAny(cqrs.NewCommand[UpdateCodePayload, any](UpdateCode, UpdateCodePayload{
				Color: "teal-500",
			}, EntityName, "code-456", nil)),
			expectErr: false,
			expectEvent: cqrs.ToAny(cqrs.NewDomainEvent[UpdatedCodePayload, any](UpdatedCode, UpdatedCodePayload{
				Color: "teal-500",
			}, EntityName, "code-456", nil)),
		},
		{
			name: "UpdateCode with reasoning only",
			input: cqrs.ToAny(cqrs.NewCommand[UpdateCodePayload, any](UpdateCode, UpdateCodePayload{
				Reasoning: "Comprehensive climate coverage",
			}, EntityName, "code-789", nil)),
			expectErr: false,
			expectEvent: cqrs.ToAny(cqrs.NewDomainEvent[UpdatedCodePayload, any](UpdatedCode, UpdatedCodePayload{
				Reasoning: "Comprehensive climate coverage",
			}, EntityName, "code-789", nil)),
		},
		{
			name:        "DeleteCode",
			input:       cqrs.ToAny(cqrs.NewCommand[any, any](DeleteCode, nil, EntityName, "code-999", nil)),
			expectErr:   false,
			expectEvent: cqrs.ToAny(cqrs.NewDomainEvent[any, any](DeletedCode, nil, EntityName, "code-999", nil)),
		},
		{
			name: "CreateCode with missing Color and Reasoning",
			input: cqrs.ToAny(cqrs.NewCommand[CreateCodePayload, any](CreateCode, CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:incomplete",
			}, EntityName, "", nil)),
			expectErr: true,
		},
		{
			name: "CreateCode with invalid slug (no colon)",
			input: cqrs.ToAny(cqrs.NewCommand[CreateCodePayload, any](CreateCode, CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topicclimate",
				Color:     "blue-500",
				Reasoning: "Invalid format",
			}, EntityName, "", nil)),
			expectErr: true,
		},
		{
			name: "CreateCode with invalid slug (uppercase)",
			input: cqrs.ToAny(cqrs.NewCommand[CreateCodePayload, any](CreateCode, CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "Topic:climate",
				Color:     "blue-500",
				Reasoning: "Invalid format",
			}, EntityName, "", nil)),
			expectErr: true,
		},
		{
			name: "CreateCode with invalid slug (empty after colon)",
			input: cqrs.ToAny(cqrs.NewCommand[CreateCodePayload, any](CreateCode, CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:",
				Color:     "blue-500",
				Reasoning: "Invalid format",
			}, EntityName, "", nil)),
			expectErr: true,
		},
		{
			name: "CreateCode with invalid slug (starts with hyphen)",
			input: cqrs.ToAny(cqrs.NewCommand[CreateCodePayload, any](CreateCode, CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:-climate",
				Color:     "blue-500",
				Reasoning: "Invalid format",
			}, EntityName, "", nil)),
			expectErr: true,
		},
		{
			name: "Wrong entity type returns nil",
			input: cqrs.ToAny(cqrs.NewCommand[CreateCodePayload, any](CreateCode, CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:test",
				Color:     "blue-500",
				Reasoning: "Test",
			}, "DifferentEntity", "", nil)),
			expectErr:   false,
			expectEvent: nil,
		},
		{
			name: "Wrong message type returns nil",
			input: cqrs.ToAny(cqrs.NewDomainEvent[CreateCodePayload, any](CreateCode, CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:test",
				Color:     "blue",
				Reasoning: "Test",
			}, EntityName, "", nil)),
			expectErr:   false,
			expectEvent: nil,
		},
		{
			name: "Wrong action returns nil",
			input: cqrs.ToAny(cqrs.NewCommand[CreateCodePayload, any]("DifferentAction", CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:test",
				Color:     "blue-500",
				Reasoning: "Test",
			}, EntityName, "test-aggregate-id", nil)),
			expectErr:   false,
			expectEvent: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Router(tt.input, nil)

			if tt.expectErr {
				th.AssertNotNil(t, err, "should return error")
				return
			}

			th.AssertNil(t, err, "should not return error")
			th.AssertMessage(t, result, tt.expectEvent, "event")
		})
	}
}
