package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestProjectRouter(t *testing.T) {
	tests := []struct {
		name        string
		input       *commands.AnyMessage
		expectErr   string
		expectEvent *commands.AnyMessage
	}{
		{
			name: "CreateProject with valid payload",
			input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](project.CreateProject, project.CreateProjectPayload{
				Name: "My Research Project",
			}, project.EntityName, "", nil)),
			expectErr: "",
			expectEvent: commands.ToAny(commands.NewDomainEvent[project.CreatedProjectPayload, any](project.CreatedProject, project.CreatedProjectPayload{
				Name: "My Research Project",
			}, project.EntityName, "", nil)),
		},
		{
			name:      "CreateProject with missing Name",
			input:     commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](project.CreateProject, project.CreateProjectPayload{}, project.EntityName, "", nil)),
			expectErr: "validation failed: Name is required",
		},
		{
			name: "Wrong entity type returns nil",
			input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](project.CreateProject, project.CreateProjectPayload{
				Name: "Test Project",
			}, "DifferentEntity", "", nil)),
			expectErr:   "",
			expectEvent: nil,
		},
		{
			name: "Wrong action returns nil",
			input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any]("DifferentAction", project.CreateProjectPayload{
				Name: "Test Project",
			}, project.EntityName, "test-aggregate-id", nil)),
			expectErr:   "",
			expectEvent: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Router(tt.input, nil)

			th.AssertError(t, err, tt.expectErr, "error")
			if tt.expectErr == "" {
				th.AssertMessage(t, result, tt.expectEvent, "event")
			}
		})
	}
}
