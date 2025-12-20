package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/templates"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"testing"
)

func TestBuildCreateDocumentCommand(t *testing.T) {
	tests := []struct {
		name      string
		projectID string
		dd        templates.DefaultDocument
		wantName  string
	}{
		{
			name:      "codebook document",
			projectID: "proj-1",
			dd:        templates.DefaultDocument{Name: "Codebook", Content: "test"},
			wantName:  "Codebook",
		},
		{
			name:      "memo document",
			projectID: "proj-2",
			dd:        templates.DefaultDocument{Name: "Notes", Content: "memo content"},
			wantName:  "Notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := buildCreateDocumentCommand(tt.projectID, tt.dd, domain_helpers.TestActor())

			th.AssertEqual(t, cmd.Action, document.CreateDocument, "action")
			th.AssertEqual(t, cmd.AggregateType, document.EntityName, "entity")

			payload := cmd.Payload.(document.CreateDocumentPayload)
			th.AssertEqual(t, payload.ProjectID, tt.projectID, "projectID")
			th.AssertEqual(t, payload.Name, tt.wantName, "name")
		})
	}
}

func TestCreateDefaultDocuments(t *testing.T) {
	message := project.CreatedProjectEvent("proj-1")
	result := createDefaultDocuments(message, project.CreatedProjectPayload{})

	assertDefaultDocumentCommands(t, result, "proj-1")
}

func assertDefaultDocumentCommands(t *testing.T, cmds []*commands.AnyMessage, expectedProjectID string) {
	t.Helper()
	defaultDocs := templates.DefaultDocuments()

	if len(cmds) != len(defaultDocs) {
		t.Fatalf("expected %d commands, got %d", len(defaultDocs), len(cmds))
	}

	for i, cmd := range cmds {
		th.AssertEqual(t, cmd.Action, document.CreateDocument, "action")
		th.AssertEqual(t, cmd.AggregateType, document.EntityName, "entity")

		payload := cmd.Payload.(document.CreateDocumentPayload)
		th.AssertEqual(t, payload.ProjectID, expectedProjectID, "projectID")
		th.AssertEqual(t, payload.Name, defaultDocs[i].Name, "name")
	}
}
