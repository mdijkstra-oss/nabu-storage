package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	rh "hermes-relay/internal/lib/test-helpers/registry-helpers"
	"testing"
)

func TestExternalEventHandlers(t *testing.T) {
	setupEvents := []*commands.AnyMessage{
		commands.ToAny(commands.NewDomainEvent[project.CreatedProjectPayload, any](
			project.CreatedProject,
			project.CreatedProjectPayload{Name: "Test Project"},
			project.EntityName,
			"project-1",
			nil,
		)),
	}

	tests := []rh.RouterTestCase{
		{
			Name: "OnFileCreated produces AddedFileToProject event",
			Input: commands.ToAny(commands.NewDomainEvent[file.CreatedFilePayload, any](
				file.CreatedFile,
				file.CreatedFilePayload{
					BaseFile: file.BaseFile{
						ID:        "file-1",
						ProjectID: "project-1",
						Name:      "test.md",
						Attributes: file.Attributes{
							Title: "Test File",
						},
					},
					Content: "content",
				},
				file.EntityName,
				"file-1",
				nil,
			)),
			ExpectErr:   "",
			ExpectEvent: nil,
			ExpectPublished: []*commands.AnyMessage{
				commands.ToAny(commands.NewDomainEvent[project.AddedFileToProjectPayload, any](
					project.AddedFileToProject,
					project.AddedFileToProjectPayload{
						FileID:    "file-1",
						ProjectID: "project-1",
					},
					project.EntityName,
					"project-1",
					nil,
				)),
			},
		},
		{
			Name: "OnCodeCreated produces AddedCodeToProject event",
			Input: commands.ToAny(commands.NewDomainEvent[code.CreatedCodePayload, any](
				code.CreatedCode,
				code.CreatedCodePayload{
					ProjectID: "project-1",
					Slug:      "theme:nature",
					Color:     "#00FF00",
					Reasoning: "Nature themed",
				},
				code.EntityName,
				"code-1",
				nil,
			)),
			ExpectErr:   "",
			ExpectEvent: nil,
			ExpectPublished: []*commands.AnyMessage{
				commands.ToAny(commands.NewDomainEvent[project.AddedCodeToProjectPayload, any](
					project.AddedCodeToProject,
					project.AddedCodeToProjectPayload{
						CodeID:    "code-1",
						ProjectID: "project-1",
					},
					project.EntityName,
					"project-1",
					nil,
				)),
			},
		},
		{
			Name: "OnCodeDeleted produces RemovedCodeFromProject event",
			Input: commands.ToAny(commands.NewDomainEvent[code.DeletedCodePayload, any](
				code.DeletedCode,
				code.DeletedCodePayload{
					ProjectID: "project-1",
				},
				code.EntityName,
				"code-1",
				nil,
			)),
			ExpectErr:   "",
			ExpectEvent: nil,
			ExpectPublished: []*commands.AnyMessage{
				commands.ToAny(commands.NewDomainEvent[project.RemovedCodeFromProjectPayload, any](
					project.RemovedCodeFromProject,
					project.RemovedCodeFromProjectPayload{
						CodeID:    "code-1",
						ProjectID: "project-1",
					},
					project.EntityName,
					"project-1",
					nil,
				)),
			},
		},
		{
			Name: "Unrelated file event returns nil",
			Input: commands.ToAny(commands.NewDomainEvent[file.CodedFilePayload, any](
				file.CodedFile,
				file.CodedFilePayload{
					Actions: []file.CodingAction{
						{
							CodeSlug: "theme:nature",
							CodeID:   "code-1",
							Action:   file.SetCoding,
							Sections: []file.CodedSectionAttributes{{Text: "sample text"}},
							ChunkID:  "chunk-1",
						},
					},
				},
				file.EntityName,
				"file-1",
				nil,
			)),
			ExpectErr:       "",
			ExpectEvent:     nil,
			ExpectPublished: nil,
		},
		{
			Name: "Unrelated code event returns nil",
			Input: commands.ToAny(commands.NewDomainEvent[code.UpdatedCodePayload, any](
				code.UpdatedCode,
				code.UpdatedCodePayload{
					Color: "#FF0000",
				},
				code.EntityName,
				"code-1",
				nil,
			)),
			ExpectErr:       "",
			ExpectEvent:     nil,
			ExpectPublished: nil,
		},
	}

	rh.RunRouterTests(t, setupEvents, tests, NewRouter)
}
