package domain

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	rh "hermes-relay/internal/lib/test-helpers/router-helpers"
	"hermes-relay/internal/lib/utils"
	"testing"
)

var (
	healthyProjectID = utils.NewID()
	healthyCodeID    = utils.NewID()
	healthyFileID    = utils.NewID()
)

var setupCommands = []*commands.AnyMessage{
	commands.ToAny(commands.NewDomainEvent[project.CreatedProjectPayload, any](
		project.CreatedProject,
		project.CreatedProjectPayload{
			Name:        "Healthy Project",
			Description: "Test project",
		},
		project.EntityName,
		healthyProjectID,
		domain_helpers.TestActor(),
		nil,
	)),
	commands.ToAny(commands.NewDomainEvent[code.CreatedCodePayload, any](
		code.CreatedCode,
		code.CreatedCodePayload{
			ProjectID:  healthyProjectID,
			Slug:       "topic:existing-code",
			Color:      "green",
			Definition: "Existing code",
		},
		code.EntityName,
		healthyCodeID,
		domain_helpers.TestActor(),
		nil,
	)),
	commands.ToAny(commands.NewDomainEvent[file.CreatedFilePayload, any](
		file.CreatedFile,
		file.CreatedFilePayload{
			FileData: file.FileData{
				ProjectID: healthyProjectID,
				Name:      "Existing File",
				Type:      file.FileTypeCorpus,
				Locked:    true,
			},
			Chunks: []file.Chunk{{ID: "1", Content: "Test content"}},
		},
		file.EntityName,
		healthyFileID,
		domain_helpers.TestActor(),
		nil,
	)),
}

func TestCommandRouter(t *testing.T) {
	tests := []rh.RouterTestCase{
		{
			Name: "Project creation succeeds with valid payload",
			Input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](
				project.CreateProject,
				project.CreateProjectPayload{Name: "New Project", Description: "Test"},
				project.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[project.CreatedProjectPayload, any](
				project.CreatedProject,
				project.CreatedProjectPayload{
					Name:        "New Project",
					Description: "Test",
				},
				project.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
		},
		{
			Name: "Project creation fails with invalid payload (missing Name)",
			Input: commands.ToAny(commands.NewCommand[project.CreateProjectPayload, any](
				project.CreateProject,
				project.CreateProjectPayload{Name: "", Description: "Test"},
				project.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "validation failed: Name is required",
		},
		{
			Name: "Code creation succeeds on healthy project",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](
				code.CreateCode,
				code.CreateCodePayload{
					ProjectID:  healthyProjectID,
					Slug:       "topic:new-code",
					Color:      "blue",
					Definition: "Test code",
				},
				code.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[code.CreatedCodePayload, any](
				code.CreatedCode,
				code.CreatedCodePayload{
					ProjectID:  healthyProjectID,
					Slug:       "topic:new-code",
					Color:      "blue",
					Definition: "Test code",
				},
				code.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
		},
		{
			Name: "Code creation fails on non-existent project",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](
				code.CreateCode,
				code.CreateCodePayload{
					ProjectID:  utils.NewID(),
					Slug:       "topic:orphan",
					Color:      "gray",
					Definition: "No parent",
				},
				code.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "validation failed: ProjectID not found",
		},
		{
			Name: "Code update succeeds on healthy code in healthy project",
			Input: commands.ToAny(commands.NewCommand[code.UpdateCodePayload, any](
				code.UpdateCode,
				code.UpdateCodePayload{
					Color:      "teal",
					Definition: "Updated",
				},
				code.EntityName,
				healthyCodeID,
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[code.UpdatedCodePayload, any](
				code.UpdatedCode,
				code.UpdatedCodePayload{
					Color:      "teal",
					Definition: "Updated",
				},
				code.EntityName,
				healthyCodeID,
				domain_helpers.TestActor(),
				nil,
			)),
		},
		{
			Name: "File creation succeeds on healthy project",
			Input: commands.ToAny(commands.NewCommand[file.CreateFilePayload, any](
				file.CreateFile,
				file.CreateFilePayload{
					ProjectID: healthyProjectID,
					Name:      "New File",
					Content:   "Content",
				},
				file.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.CreatedFilePayload, any](
				file.CreatedFile,
				file.CreatedFilePayload{
					FileData: file.FileData{
						ProjectID: healthyProjectID,
						Name:      "New File",
						Type:      file.FileTypeCorpus,
						Locked:    true,
					},
					Chunks: []file.Chunk{{ID: "1", Content: "Content\n", Codes: []file.CodedSection{}}},
				},
				file.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
		},
		{
			Name: "File update succeeds on healthy file in healthy project",
			Input: commands.ToAny(commands.NewCommand[file.UpdateFilePayload, any](
				file.UpdateFile,
				file.UpdateFilePayload{
					Name:        "Updated File",
					Description: "Updated",
				},
				file.EntityName,
				healthyFileID,
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[file.UpdatedFilePayload, any](
				file.UpdatedFile,
				file.UpdatedFilePayload{
					Name:        "Updated File",
					Description: "Updated",
				},
				file.EntityName,
				healthyFileID,
				domain_helpers.TestActor(),
				nil,
			)),
		},
		{
			Name: "Code creation fails with invalid slug (no colon)",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](
				code.CreateCode,
				code.CreateCodePayload{
					ProjectID:  healthyProjectID,
					Slug:       "invalid-slug",
					Color:      "blue",
					Definition: "Bad slug",
				},
				code.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			Name: "Code creation fails with duplicate slug",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](
				code.CreateCode,
				code.CreateCodePayload{
					ProjectID:  healthyProjectID,
					Slug:       "topic:existing-code",
					Color:      "blue",
					Definition: "Duplicate",
				},
				code.EntityName,
				"",
				domain_helpers.TestActor(),
				nil,
			)),
			ExpectErr: "validation failed: slug already in use",
		},
	}

	rh.RunRouterTests(t, setupCommands, tests, NewCommandRouter)
}
