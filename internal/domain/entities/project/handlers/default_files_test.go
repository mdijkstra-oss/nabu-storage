package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/templates"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"testing"
)

func TestBuildCreateFileCommand(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		df        templates.DefaultFile
		wantType  file.FileType
		wantName  string
	}{
		{
			name:      "codebook file",
			projectID: "proj-1",
			df:        templates.DefaultFile{Name: "Codebook", Type: file.FileTypeCodebook, Content: "test"},
			wantType:  file.FileTypeCodebook,
			wantName:  "Codebook",
		},
		{
			name:      "memo file",
			projectID: "proj-2",
			df:        templates.DefaultFile{Name: "Notes", Type: file.FileTypeMemo, Content: "memo content"},
			wantType:  file.FileTypeMemo,
			wantName:  "Notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildCreateFileCommand(tt.projectID, tt.df, domain_helpers.TestActor())

			th.AssertEqual(t, cmd.Action, file.CreateFile, "action")
			th.AssertEqual(t, cmd.AggregateType, file.EntityName, "entity")

			payload := cmd.Payload.(file.CreateFilePayload)
			th.AssertEqual(t, payload.ProjectID, tt.projectID, "projectID")
			th.AssertEqual(t, payload.Type, tt.wantType, "type")
			th.AssertEqual(t, payload.Name, tt.wantName, "name")
		})
	}
}

func TestCreateDefaultFiles(t *testing.T) {
	message := project.CreatedProjectEvent("proj-1")
	result := createDefaultFiles(message, project.CreatedProjectPayload{})

	assertDefaultFileCommands(t, result, "proj-1")
}

func assertDefaultFileCommands(t *testing.T, cmds []*commands.AnyMessage, expectedProjectID string) {
	t.Helper()
	defaultFiles := templates.DefaultFiles()

	if len(cmds) != len(defaultFiles) {
		t.Fatalf("expected %d commands, got %d", len(defaultFiles), len(cmds))
	}

	for i, cmd := range cmds {
		th.AssertEqual(t, cmd.Action, file.CreateFile, "action")
		th.AssertEqual(t, cmd.AggregateType, file.EntityName, "entity")

		payload := cmd.Payload.(file.CreateFilePayload)
		th.AssertEqual(t, payload.ProjectID, expectedProjectID, "projectID")
		th.AssertEqual(t, payload.Type, defaultFiles[i].Type, "type")
		th.AssertEqual(t, payload.Name, defaultFiles[i].Name, "name")
	}
}
