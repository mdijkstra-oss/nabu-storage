package projectview

import (
	"hermes-relay/internal/cqrs"
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
			name: "AddedFileToProject appends to FileIDs",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{},
			},
			event: newProjectEvent("project-1", project.AddedFileToProject, &project.AddedFileToProjectPayload{
				FileID:    "file-1",
				ProjectID: "project-1",
			}),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{"file-1"},
			},
		},
		{
			name: "AddedFileToProject with multiple files",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{"file-1"},
			},
			event: newProjectEvent("project-1", project.AddedFileToProject, &project.AddedFileToProjectPayload{
				FileID:    "file-2",
				ProjectID: "project-1",
			}),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{"file-1", "file-2"},
			},
		},
		{
			name: "AddedCodeToProject appends to CodeIDs",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{},
			},
			event: newProjectEvent("project-1", project.AddedCodeToProject, &project.AddedCodeToProjectPayload{
				CodeID:    "code-1",
				ProjectID: "project-1",
			}),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1"},
				FileIDs: []string{},
			},
		},
		{
			name: "AddedCodeToProject with multiple codes",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1"},
				FileIDs: []string{},
			},
			event: newProjectEvent("project-1", project.AddedCodeToProject, &project.AddedCodeToProjectPayload{
				CodeID:    "code-2",
				ProjectID: "project-1",
			}),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1", "code-2"},
				FileIDs: []string{},
			},
		},
		{
			name: "RemovedCodeFromProject removes from CodeIDs",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1", "code-2", "code-3"},
				FileIDs: []string{},
			},
			event: newProjectEvent("project-1", project.RemovedCodeFromProject, &project.RemovedCodeFromProjectPayload{
				CodeID:    "code-2",
				ProjectID: "project-1",
			}),
			expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1", "code-3"},
				FileIDs: []string{},
			},
		},
		{
			name: "RemovedCodeFromProject on non-existent code is safe",
			initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1"},
				FileIDs: []string{},
			},
			event: newProjectEvent("project-1", project.RemovedCodeFromProject, &project.RemovedCodeFromProjectPayload{
				CodeID:    "code-999",
				ProjectID: "project-1",
			}),
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
			event: newProjectEvent("project-1", project.RemovedCodeFromProject, &project.RemovedCodeFromProjectPayload{
				CodeID:    "code-1",
				ProjectID: "project-1",
			}),
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
	return cqrs.ToAny(cqrs.NewDomainEvent[any, any](action, payload, project.EntityName, aggregateID, nil))
}
