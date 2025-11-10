package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
	rh "hermes-relay/internal/lib/test-helpers/router-helpers"
	"testing"
)

var (
	extTestProjectID = utils.NewID()
	extTestFileID    = utils.NewID()
	extTestCodeID    = utils.NewID()
)

func TestExternalEventHandlers(t *testing.T) {
	setupEvents := []*commands.AnyMessage{
		commands.ToAny(commands.NewDomainEvent[project.CreatedProjectPayload, any](
			project.CreatedProject,
			project.CreatedProjectPayload{Name: "Test Project"},
			project.EntityName,
			extTestProjectID,
			nil,
		)),
	}

	tests := []rh.RouterTestCase{
		{
			Name: "OnFileCreated produces AddedFileToProject event",
			Input: commands.ToAny(commands.NewDomainEvent[file.CreatedFilePayload, any](
				file.CreatedFile,
				file.CreatedFilePayload{
					CreateFilePayload: file.CreateFilePayload{
						ProjectID: extTestProjectID,
						Name:      "test.md",
						Content:   "content",
					},
					Type:   file.FileTypeSource,
					Locked: true,
				},
				file.EntityName,
				extTestFileID,
				nil,
			)),
			ExpectErr:   "",
			ExpectEvent: nil,
			ExpectPublished: []*commands.AnyMessage{
				commands.ToAny(commands.NewDomainEvent[project.AddedFileToProjectPayload, any](
					project.AddedFileToProject,
					project.AddedFileToProjectPayload{
						FileID:    extTestFileID,
						ProjectID: extTestProjectID,
					},
					project.EntityName,
					extTestProjectID,
					nil,
				)),
			},
		},
		{
			Name: "OnCodeCreated produces AddedCodeToProject event",
			Input: commands.ToAny(commands.NewDomainEvent[code.CreatedCodePayload, any](
				code.CreatedCode,
				code.CreatedCodePayload{
					ProjectID: extTestProjectID,
					Slug:      "theme:nature",
					Color:     "#00FF00",
					Reasoning: "Nature themed",
				},
				code.EntityName,
				extTestCodeID,
				nil,
			)),
			ExpectErr:   "",
			ExpectEvent: nil,
			ExpectPublished: []*commands.AnyMessage{
				commands.ToAny(commands.NewDomainEvent[project.AddedCodeToProjectPayload, any](
					project.AddedCodeToProject,
					project.AddedCodeToProjectPayload{
						CodeID:    extTestCodeID,
						ProjectID: extTestProjectID,
					},
					project.EntityName,
					extTestProjectID,
					nil,
				)),
			},
		},
		{
			Name: "OnCodeDeleted produces RemovedCodeFromProject event",
			Input: commands.ToAny(commands.NewDomainEvent[code.DeletedCodePayload, any](
				code.DeletedCode,
				code.DeletedCodePayload{
					ProjectID: extTestProjectID,
				},
				code.EntityName,
				extTestCodeID,
				nil,
			)),
			ExpectErr:   "",
			ExpectEvent: nil,
			ExpectPublished: []*commands.AnyMessage{
				commands.ToAny(commands.NewDomainEvent[project.RemovedCodeFromProjectPayload, any](
					project.RemovedCodeFromProject,
					project.RemovedCodeFromProjectPayload{
						CodeID:    extTestCodeID,
						ProjectID: extTestProjectID,
					},
					project.EntityName,
					extTestProjectID,
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
							CodeID:   extTestCodeID,
							Action:   file.SetCoding,
							Sections: []file.CodedSectionAttributes{{Text: "sample text"}},
							ChunkID:  "chunk-1",
						},
					},
				},
				file.EntityName,
				extTestFileID,
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
				extTestCodeID,
				nil,
			)),
			ExpectErr:       "",
			ExpectEvent:     nil,
			ExpectPublished: nil,
		},
	}

	rh.RunRouterTests(t, setupEvents, tests, NewRouter)
}
