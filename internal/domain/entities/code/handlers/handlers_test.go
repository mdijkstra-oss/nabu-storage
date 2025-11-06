package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
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
	tests := []rh.RouterTestCase{
		{
			Name: "CreateCode with valid payload",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:economy",
				Color:     "green-500",
				Reasoning: "Economic topics",
			}, code.EntityName, "", nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[code.CreatedCodePayload, any](code.CreatedCode, code.CreatedCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:economy",
				Color:     "green-500",
				Reasoning: "Economic topics",
			}, code.EntityName, "", nil)),
		},
		{
			Name: "UpdateCode with both fields for existing code",
			Input: commands.ToAny(commands.NewCommand[code.UpdateCodePayload, any](code.UpdateCode, code.UpdateCodePayload{
				Color:     "emerald-600",
				Reasoning: "Updated climate coverage",
			}, code.EntityName, "code-existing-1", nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[code.UpdatedCodePayload, any](code.UpdatedCode, code.UpdatedCodePayload{
				Color:     "emerald-600",
				Reasoning: "Updated climate coverage",
			}, code.EntityName, "code-existing-1", nil)),
		},
		{
			Name: "UpdateCode for non-existent code fails",
			Input: commands.ToAny(commands.NewCommand[code.UpdateCodePayload, any](code.UpdateCode, code.UpdateCodePayload{
				Color: "red-500",
			}, code.EntityName, "code-nonexistent", nil)),
			ExpectErr: "validation failed: ProjectID not found",
		},
		{
			Name: "DeleteCode",
			Input: commands.ToAny(commands.NewCommand[code.DeleteCodeData, any](code.DeleteCode, code.DeleteCodeData{
				ProjectID: "project-1",
			}, code.EntityName, "code-999", nil)),
			ExpectErr:   "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[any, any](code.DeletedCode, nil, code.EntityName, "code-999", nil)),
		},
		{
			Name: "CreateCode with missing Color and Reasoning",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:incomplete",
			}, code.EntityName, "", nil)),
			ExpectErr: "validation failed: Color is required, Reasoning is required",
		},
		{
			Name: "CreateCode with invalid slug (no colon)",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topicclimate",
				Color:     "blue-500",
				Reasoning: "Invalid format",
			}, code.EntityName, "", nil)),
			ExpectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			Name: "CreateCode with invalid slug (uppercase)",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "Topic:climate",
				Color:     "blue-500",
				Reasoning: "Invalid format",
			}, code.EntityName, "", nil)),
			ExpectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			Name: "CreateCode with invalid slug (empty after colon)",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:",
				Color:     "blue-500",
				Reasoning: "Invalid format",
			}, code.EntityName, "", nil)),
			ExpectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			Name: "CreateCode with invalid slug (starts with hyphen)",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:-climate",
				Color:     "blue-500",
				Reasoning: "Invalid format",
			}, code.EntityName, "", nil)),
			ExpectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			Name: "Wrong entity type returns nil",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:test",
				Color:     "blue-500",
				Reasoning: "Test",
			}, "DifferentEntity", "", nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
		{
			Name: "Wrong action returns nil",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any]("DifferentAction", code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:test",
				Color:     "blue-500",
				Reasoning: "Test",
			}, code.EntityName, "test-aggregate-id", nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
		{
			Name: "CreateCode with duplicate slug fails",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "blue-500",
				Reasoning: "Duplicate slug",
			}, code.EntityName, "", nil)),
			ExpectErr: "validation failed: slug already in use",
		},
		{
			Name: "MergeCodes with valid source and target",
			Input: commands.ToAny(commands.NewCommand[code.MergeCodesPayload, any](code.MergeCodes, code.MergeCodesPayload{
				SourceID: "code-existing-1",
				TargetID: "code-existing-2",
			}, code.EntityName, "code-existing-1", nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[code.MergedCodesPayload, any](code.MergedCodes, code.MergedCodesPayload{
				SourceID: "code-existing-1",
				TargetID: "code-existing-2",
			}, code.EntityName, "code-existing-1", nil)),
		},
		{
			Name: "MergeCodes with non-existent source fails",
			Input: commands.ToAny(commands.NewCommand[code.MergeCodesPayload, any](code.MergeCodes, code.MergeCodesPayload{
				SourceID: "code-nonexistent",
				TargetID: "code-existing-2",
			}, code.EntityName, "code-existing-2", nil)),
			ExpectErr: "validation failed: source_id not found",
		},
		{
			Name: "MergeCodes with non-existent target fails",
			Input: commands.ToAny(commands.NewCommand[code.MergeCodesPayload, any](code.MergeCodes, code.MergeCodesPayload{
				SourceID: "code-existing-1",
				TargetID: "code-nonexistent",
			}, code.EntityName, "code-existing-1", nil)),
			ExpectErr: "validation failed: target_id not found",
		},
		{
			Name: "MergeCodes with same source and target fails",
			Input: commands.ToAny(commands.NewCommand[code.MergeCodesPayload, any](code.MergeCodes, code.MergeCodesPayload{
				SourceID: "code-existing-1",
				TargetID: "code-existing-1",
			}, code.EntityName, "code-existing-1", nil)),
			ExpectErr: "validation failed: source_id cannot merge with itself",
		},
	}

	rh.RunRouterTests(t, cmds, tests, NewRouter)
}
