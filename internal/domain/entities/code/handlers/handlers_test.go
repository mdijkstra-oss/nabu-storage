package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	th "hermes-relay/internal/lib/test-helpers"
	rh "hermes-relay/internal/lib/test-helpers/registry-helpers"
	"testing"
)

var cmds = []*commands.AnyMessage{
	// Create project
	commands.ToAny(commands.NewDomainEvent[any, any](
		"CreatedProject",
		map[string]any{"name": "Test Project"},
		"Project",
		"project-1",
		nil,
	)),
	// Create existing codes
	commands.ToAny(commands.NewDomainEvent[code.CreatedCodePayload, any](
		code.CreatedCode,
		code.CreatedCodePayload{
			ProjectID: "project-1",
			Slug:      "topic:climate",
			Color:     "blue-500",
			Reasoning: "Climate topics",
		},
		code.EntityName,
		"code-existing-1",
		nil,
	)),
	commands.ToAny(commands.NewDomainEvent[code.CreatedCodePayload, any](
		code.CreatedCode,
		code.CreatedCodePayload{
			ProjectID: "project-1",
			Slug:      "topic:health",
			Color:     "green-500",
			Reasoning: "Health topics",
		},
		code.EntityName,
		"code-existing-2",
		nil,
	)),
}

func TestCodeRouter(t *testing.T) {
	tests := []struct {
		name        string
		input       *commands.AnyMessage
		expectErr   string
		expectEvent *commands.AnyMessage
	}{
		{
			name: "CreateCode with valid payload",
			input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:economy",
				Color:     "green-500",
				Reasoning: "Economic topics",
			}, code.EntityName, "", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[code.CreatedCodePayload, any](code.CreatedCode, code.CreatedCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:economy",
				Color:     "green-500",
				Reasoning: "Economic topics",
			}, code.EntityName, "", nil)),
		},
		{
			name: "UpdateCode with both fields for existing code",
			input: commands.ToAny(commands.NewCommand[code.UpdateCodePayload, any](code.UpdateCode, code.UpdateCodePayload{
				Color:     "emerald-600",
				Reasoning: "Updated climate coverage",
			}, code.EntityName, "code-existing-1", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[code.UpdatedCodePayload, any](code.UpdatedCode, code.UpdatedCodePayload{
				Color:     "emerald-600",
				Reasoning: "Updated climate coverage",
			}, code.EntityName, "code-existing-1", nil)),
		},
		{
			name: "UpdateCode for non-existent code fails",
			input: commands.ToAny(commands.NewCommand[code.UpdateCodePayload, any](code.UpdateCode, code.UpdateCodePayload{
				Color: "red-500",
			}, code.EntityName, "code-nonexistent", nil)),
			expectErr: "validation failed",
		},
		{
			name: "DeleteCode",
			input: commands.ToAny(commands.NewCommand[code.DeleteCodeData, any](code.DeleteCode, code.DeleteCodeData{
				ProjectID: "project-1",
			}, code.EntityName, "code-999", nil)),
			expectErr:   "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[any, any](code.DeletedCode, nil, code.EntityName, "code-999", nil)),
		},
		{
			name: "CreateCode with missing Color and Reasoning",
			input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:incomplete",
			}, code.EntityName, "", nil)),
			expectErr: "validation failed: Color is required, Reasoning is required",
		},
		{
			name: "CreateCode with invalid slug (no colon)",
			input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topicclimate",
				Color:     "blue-500",
				Reasoning: "Invalid format",
			}, code.EntityName, "", nil)),
			expectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			name: "CreateCode with invalid slug (uppercase)",
			input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "Topic:climate",
				Color:     "blue-500",
				Reasoning: "Invalid format",
			}, code.EntityName, "", nil)),
			expectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			name: "CreateCode with invalid slug (empty after colon)",
			input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:",
				Color:     "blue-500",
				Reasoning: "Invalid format",
			}, code.EntityName, "", nil)),
			expectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			name: "CreateCode with invalid slug (starts with hyphen)",
			input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:-climate",
				Color:     "blue-500",
				Reasoning: "Invalid format",
			}, code.EntityName, "", nil)),
			expectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			name: "Wrong entity type returns nil",
			input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:test",
				Color:     "blue-500",
				Reasoning: "Test",
			}, "DifferentEntity", "", nil)),
			expectErr:   "",
			expectEvent: nil,
		},
		{
			name: "Wrong message type returns nil",
			input: commands.ToAny(commands.NewDomainEvent[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:test",
				Color:     "blue",
				Reasoning: "Test",
			}, code.EntityName, "", nil)),
			expectErr:   "",
			expectEvent: nil,
		},
		{
			name: "Wrong action returns nil",
			input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any]("DifferentAction", code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:test",
				Color:     "blue-500",
				Reasoning: "Test",
			}, code.EntityName, "test-aggregate-id", nil)),
			expectErr:   "",
			expectEvent: nil,
		},
		{
			name: "CreateCode with duplicate slug fails",
			input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "blue-500",
				Reasoning: "Duplicate slug",
			}, code.EntityName, "", nil)),
			expectErr: "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := rh.NewTestRegistry(cmds)

			result, err := NewRouter(reg)(tt.input, nil)

			th.AssertError(t, err, tt.expectErr, "error")
			if tt.expectErr == "" {
				th.AssertMessage(t, result, tt.expectEvent, "event")
			}
		})
	}
}
