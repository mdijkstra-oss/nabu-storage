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
		expectErr   bool
		expectEvent *cqrs.AnyMessage
	}{
		{
			name: "CreateProject with valid payload",
			input: cqrs.ToAny(cqrs.NewCommand[CreateProjectPayload, any](CreateProject, CreateProjectPayload{
				Name: "My Research Project",
			}, EntityName, "", nil)),
			expectErr: false,
			expectEvent: cqrs.ToAny(cqrs.NewDomainEvent[CreatedProjectPayload, any](CreatedProject, CreatedProjectPayload{
				Name: "My Research Project",
			}, EntityName, "", nil)),
		},
		{
			name: "CreateProject with missing Name",
			input: cqrs.ToAny(cqrs.NewCommand[CreateProjectPayload, any](CreateProject, CreateProjectPayload{}, EntityName, "", nil)),
			expectErr: true,
		},
		{
			name: "Wrong entity type returns nil",
			input: cqrs.ToAny(cqrs.NewCommand[CreateProjectPayload, any](CreateProject, CreateProjectPayload{
				Name: "Test Project",
			}, "DifferentEntity", "", nil)),
			expectErr:   false,
			expectEvent: nil,
		},
		{
			name: "Wrong message type returns nil",
			input: cqrs.ToAny(cqrs.NewDomainEvent[CreateProjectPayload, any](CreateProject, CreateProjectPayload{
				Name: "Test Project",
			}, EntityName, "", nil)),
			expectErr:   false,
			expectEvent: nil,
		},
		{
			name: "Wrong action returns nil",
			input: cqrs.ToAny(cqrs.NewCommand[CreateProjectPayload, any]("DifferentAction", CreateProjectPayload{
				Name: "Test Project",
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
