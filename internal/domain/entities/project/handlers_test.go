package project

import (
	"hermes-relay/internal/cqrs"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestProjectRouter(t *testing.T) {
	tests := []struct {
		name        string
		input       *cqrs.AnyMessage
		expectErr   string
		expectEvent *cqrs.AnyMessage
	}{
		{
			name: "CreateProject with valid payload",
			input: cqrs.ToAny(cqrs.NewCommand[CreateProjectPayload, any](CreateProject, CreateProjectPayload{
				Name: "My Research Project",
			}, EntityName, "", nil)),
			expectErr: "",
			expectEvent: cqrs.ToAny(cqrs.NewDomainEvent[CreatedProjectPayload, any](CreatedProject, CreatedProjectPayload{
				Name: "My Research Project",
			}, EntityName, "", nil)),
		},
		{
			name: "CreateProject with missing Name",
			input: cqrs.ToAny(cqrs.NewCommand[CreateProjectPayload, any](CreateProject, CreateProjectPayload{}, EntityName, "", nil)),
			expectErr: "validation failed: Name is required",
		},
		{
			name: "Wrong entity type returns nil",
			input: cqrs.ToAny(cqrs.NewCommand[CreateProjectPayload, any](CreateProject, CreateProjectPayload{
				Name: "Test Project",
			}, "DifferentEntity", "", nil)),
			expectErr:   "",
			expectEvent: nil,
		},
		{
			name: "Wrong message type returns nil",
			input: cqrs.ToAny(cqrs.NewDomainEvent[CreateProjectPayload, any](CreateProject, CreateProjectPayload{
				Name: "Test Project",
			}, EntityName, "", nil)),
			expectErr:   "",
			expectEvent: nil,
		},
		{
			name: "Wrong action returns nil",
			input: cqrs.ToAny(cqrs.NewCommand[CreateProjectPayload, any]("DifferentAction", CreateProjectPayload{
				Name: "Test Project",
			}, EntityName, "test-aggregate-id", nil)),
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
