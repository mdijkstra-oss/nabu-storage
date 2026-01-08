package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"testing"
)

func TestEnsureExpectedVersion(t *testing.T) {
	createProject := domain_helpers.NewDomainEvent(project.EntityName, "proj-1", project.CreatedProject, &project.CreatedProjectPayload{Name: "Test"})
	updateProject := domain_helpers.NewDomainEvent(project.EntityName, "proj-1", project.UpdatedProject, &project.UpdatedProjectPayload{Name: "Updated"})
	createDocument := domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.CreatedDocument, &document.CreatedDocumentPayload{ProjectID: "proj-1", Name: "test.md"})

	handlerCalled := false
	passthrough := func(msg *commands.AnyMessage, pub dispatch.PublishFunc) (*commands.AnyMessage, error) {
		handlerCalled = true
		return msg, nil
	}

	tests := []struct {
		name      string
		setup     []*commands.AnyMessage
		command   *commands.AnyMessage
		expectErr string
	}{
		{"No expected version passes", []*commands.AnyMessage{createProject}, domain_helpers.NewCommand(project.EntityName, "proj-1", project.UpdateProject, nil), ""},
		{"Project match", []*commands.AnyMessage{createProject}, domain_helpers.NewCommandWithExpectedVersion(project.EntityName, "proj-1", project.UpdateProject, nil, 1), ""},
		{"Project mismatch", []*commands.AnyMessage{createProject, updateProject}, domain_helpers.NewCommandWithExpectedVersion(project.EntityName, "proj-1", project.UpdateProject, nil, 1), "version conflict: expected 1, actual 2"},
		{"Document match", []*commands.AnyMessage{createProject, createDocument}, domain_helpers.NewCommandWithExpectedVersion(document.EntityName, "doc-1", document.UpdateDocument, nil, 1), ""},
		{"Non-existent entity passes", []*commands.AnyMessage{createProject}, domain_helpers.NewCommandWithExpectedVersion(document.EntityName, "nonexistent", document.UpdateDocument, nil, 1), ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handlerCalled = false
			store := NewStore()
			for _, evt := range tc.setup {
				projection.Apply(store, evt)
			}
			handler := EnsureExpectedVersion(store, passthrough)
			_, err := handler(tc.command, nil)

			if tc.expectErr == "" {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
				if !handlerCalled {
					t.Error("expected handler to be called")
				}
			} else {
				if err == nil || err.Error() != tc.expectErr {
					t.Errorf("expected error %q, got: %v", tc.expectErr, err)
				}
			}
		})
	}
}
