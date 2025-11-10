package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/project"
	rh "hermes-relay/internal/lib/test-helpers/router-helpers"
	"testing"
)

var cmds = []*commands.AnyMessage{}

func TestProjectRouter(t *testing.T) {
	tests := []rh.RouterTestCase{
		{
			Name: "CreateProject with valid payload",
			Input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](project.CreateProject, project.CreateProjectPayload{
				Name: "My Research Project",
			}, project.EntityName, "", nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[project.CreatedProjectPayload, any](project.CreatedProject, project.CreatedProjectPayload{
				Name: "My Research Project",
			}, project.EntityName, "", nil)),
		},
		{
			Name:      "CreateProject with missing Name",
			Input:     commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](project.CreateProject, project.CreateProjectPayload{}, project.EntityName, "", nil)),
			ExpectErr: "validation failed: Name is required",
		},
		{
			Name: "Wrong entity type returns nil",
			Input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](project.CreateProject, project.CreateProjectPayload{
				Name: "Test Project",
			}, "DifferentEntity", "", nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
		{
			Name: "Wrong action returns nil",
			Input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any]("DifferentAction", project.CreateProjectPayload{
				Name: "Test Project",
			}, project.EntityName, "test-aggregate-id", nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
	}

	rh.RunRouterTests(t, cmds, tests, NewRouter)
}
