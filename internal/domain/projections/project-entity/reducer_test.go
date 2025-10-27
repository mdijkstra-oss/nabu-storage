package projectview

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestProjectReducer(t *testing.T) {
	tests := []struct {
		name     string
		initial  *Project
		event    *cqrs.AnyMessage
		expected *Project
	}{
		{
			name:    "CreatedProject initializes empty arrays",
			initial: nil,
			event:   newProjectEvent("project-1", project.CreatedProject, &project.CreatedProjectPayload{Name: "Research"}),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{},
			},
		},
		{
			name: "FileCreated appends to FileIDs when ProjectID matches",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{},
			},
			event: newFileEvent("file-1", file.CreatedFile, &file.CreatedFilePayload{
				BaseFile: file.BaseFile{ProjectID: "project-1", Name: "doc.txt"},
			}),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{"file-1"},
			},
		},
		{
			name: "FileCreated with multiple files",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{"file-1"},
			},
			event: newFileEvent("file-2", file.CreatedFile, &file.CreatedFilePayload{
				BaseFile: file.BaseFile{ProjectID: "project-1", Name: "doc2.txt"},
			}),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{"file-1", "file-2"},
			},
		},
		{
			name: "FileCreated ignores when ProjectID doesn't match",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{},
			},
			event: newFileEvent("file-1", file.CreatedFile, &file.CreatedFilePayload{
				BaseFile: file.BaseFile{ProjectID: "project-2", Name: "doc.txt"},
			}),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{},
			},
		},
		{
			name: "CodeCreated appends to CodeIDs when ProjectID matches",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{},
			},
			event: newCodeEvent("code-1", code.CreatedCode, &code.CreatedCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green",
				Reasoning: "Climate",
			}),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1"},
				FileIDs: []string{},
			},
		},
		{
			name: "CodeCreated with multiple codes",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1"},
				FileIDs: []string{},
			},
			event: newCodeEvent("code-2", code.CreatedCode, &code.CreatedCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:policy",
				Color:     "blue",
				Reasoning: "Policy",
			}),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1", "code-2"},
				FileIDs: []string{},
			},
		},
		{
			name: "CodeCreated ignores when ProjectID doesn't match",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{},
			},
			event: newCodeEvent("code-1", code.CreatedCode, &code.CreatedCodePayload{
				ProjectID: "project-2",
				Slug:      "topic:other",
				Color:     "red",
				Reasoning: "Other",
			}),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{},
			},
		},
		{
			name: "CodeDeleted removes from CodeIDs",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1", "code-2", "code-3"},
				FileIDs: []string{},
			},
			event: newCodeEvent("code-2", code.DeletedCode, nil),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1", "code-3"},
				FileIDs: []string{},
			},
		},
		{
			name: "CodeDeleted on non-existent code is safe",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1"},
				FileIDs: []string{},
			},
			event: newCodeEvent("code-999", code.DeletedCode, nil),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1"},
				FileIDs: []string{},
			},
		},
		{
			name: "Full lifecycle: create, add files, add codes, delete code",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1", "code-2"},
				FileIDs: []string{"file-1", "file-2"},
			},
			event: newCodeEvent("code-1", code.DeletedCode, nil),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-2"},
				FileIDs: []string{"file-1", "file-2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reducer(tt.initial, tt.event)
			th.AssertEqual(t, result, tt.expected, "state after reduction")
		})
	}
}

func newProjectEvent(aggregateID string, action cqrs.Action, payload any) *cqrs.AnyMessage {
	return cqrs.ToAny(cqrs.NewDomainEvent(action, payload, project.EntityName, aggregateID, (*cqrs.AnyMessage)(nil)))
}

func newFileEvent(aggregateID string, action cqrs.Action, payload any) *cqrs.AnyMessage {
	return cqrs.ToAny(cqrs.NewDomainEvent(action, payload, file.EntityName, aggregateID, (*cqrs.AnyMessage)(nil)))
}

func newCodeEvent(aggregateID string, action cqrs.Action, payload any) *cqrs.AnyMessage {
	return cqrs.ToAny(cqrs.NewDomainEvent(action, payload, code.EntityName, aggregateID, (*cqrs.AnyMessage)(nil)))
}
