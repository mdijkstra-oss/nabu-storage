package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	rh "hermes-relay/internal/lib/test-helpers/router-helpers"
	"hermes-relay/internal/lib/utils"
	"testing"
)

var testProjectID = utils.NewID()

var cmds = []*commands.AnyMessage{}

func TestProjectRouter(t *testing.T) {
	tests := []rh.RouterTestCase{
		{
			Name: "CreateProject with valid payload defaults phase to explore",
			Input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](project.CreateProject, project.CreateProjectPayload{
				Name:        "My Research Project",
				Description: "A project for research purposes",
			}, project.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[project.CreatedProjectPayload, any](project.CreatedProject, project.CreatedProjectPayload{
				Name:        "My Research Project",
				Description: "A project for research purposes",
				Phase:       project.PhaseExplore,
			}, project.EntityName, "", domain_helpers.TestActor(), nil)),
		},
		{
			Name: "CreateProject with explicit phase",
			Input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](project.CreateProject, project.CreateProjectPayload{
				Name:  "Code Phase Project",
				Phase: project.PhaseCode,
			}, project.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[project.CreatedProjectPayload, any](project.CreatedProject, project.CreatedProjectPayload{
				Name:  "Code Phase Project",
				Phase: project.PhaseCode,
			}, project.EntityName, "", domain_helpers.TestActor(), nil)),
		},
		{
			Name: "CreateProject with invalid phase fails",
			Input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](project.CreateProject, project.CreateProjectPayload{
				Name:  "Invalid Phase Project",
				Phase: "invalid",
			}, project.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Phase failed validation (oneof)",
		},
		{
			Name:      "CreateProject with missing Name",
			Input:     commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](project.CreateProject, project.CreateProjectPayload{}, project.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Name is required",
		},
		{
			Name: "UpdateProject with valid payload",
			Input: commands.ToAny(commands.NewCommand[project.UpdateProjectPayload, any](project.UpdateProject, project.UpdateProjectPayload{
				Name:        "Updated Project Name",
				Description: "Updated description",
			}, project.EntityName, testProjectID, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[project.UpdatedProjectPayload, any](project.UpdatedProject, project.UpdatedProjectPayload{
				Name:        "Updated Project Name",
				Description: "Updated description",
			}, project.EntityName, testProjectID, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "UpdateProject with phase change",
			Input: commands.ToAny(commands.NewCommand[project.UpdateProjectPayload, any](project.UpdateProject, project.UpdateProjectPayload{
				Name:  "Project Name",
				Phase: project.PhaseAnalyze,
			}, project.EntityName, testProjectID, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[project.UpdatedProjectPayload, any](project.UpdatedProject, project.UpdatedProjectPayload{
				Name:  "Project Name",
				Phase: project.PhaseAnalyze,
			}, project.EntityName, testProjectID, domain_helpers.TestActor(), nil)),
		},
		{
			Name:      "UpdateProject with missing Name",
			Input:     commands.ToAny(commands.NewCommand[project.UpdateProjectPayload, any](project.UpdateProject, project.UpdateProjectPayload{}, project.EntityName, testProjectID, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Name is required",
		},
		{
			Name:        "DeleteProject with valid aggregate ChunkID",
			Input:       commands.ToAny(commands.NewCommand[project.DeleteProjectPayload, any](project.DeleteProject, project.DeleteProjectPayload{}, project.EntityName, testProjectID, domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[any, any](project.DeletedProject, nil, project.EntityName, testProjectID, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "Wrong entity type returns nil",
			Input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](project.CreateProject, project.CreateProjectPayload{
				Name: "Test Project",
			}, "DifferentEntity", "", domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
		{
			Name: "Wrong action returns nil",
			Input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any]("DifferentAction", project.CreateProjectPayload{
				Name: "Test Project",
			}, project.EntityName, "test-aggregate-id", domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
	}

	rh.RunRouterTests(t, cmds, tests, NewRouter)
}
