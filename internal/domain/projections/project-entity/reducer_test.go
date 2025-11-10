package projectview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/test-helpers/reducer-helpers"
	"testing"
)

func createTestProject() *Project {
	return &Project{
		ID:          "project-1",
		Name:        "Test Project",
		Description: "Test description",
		CodeIDs:     []string{"code-1"},
		FileIDs:     []string{"file-1"},
	}
}

func TestProjectReducer(t *testing.T) {
	tests := []reducer_helpers.ReducerTestCase[*Project]{
		{
			Name:    "CreatedProject initializes empty arrays",
			Initial: nil,
			Event:   newProjectEvent("project-1", project.CreatedProject, &project.CreatedProjectPayload{Name: "Research", Description: "Research project"}),
			Expected: &Project{
				ID:          "project-1",
				Name:        "Research",
				Description: "Research project",
				CodeIDs:     []string{},
				FileIDs:     []string{},
			},
		},
		{
			Name: "AddedFileToProject appends to FileIDs",
			Initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{},
			},
			Event: newProjectEvent("project-1", project.AddedFileToProject, &project.AddedFileToProjectPayload{
				FileID:    "file-1",
				ProjectID: "project-1",
			}),
			Expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{"file-1"},
			},
		},
		{
			Name: "AddedFileToProject with multiple files",
			Initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{"file-1"},
			},
			Event: newProjectEvent("project-1", project.AddedFileToProject, &project.AddedFileToProjectPayload{
				FileID:    "file-2",
				ProjectID: "project-1",
			}),
			Expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{"file-1", "file-2"},
			},
		},
		{
			Name: "AddedCodeToProject appends to CodeIDs",
			Initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{},
				FileIDs: []string{},
			},
			Event: newProjectEvent("project-1", project.AddedCodeToProject, &project.AddedCodeToProjectPayload{
				CodeID:    "code-1",
				ProjectID: "project-1",
			}),
			Expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1"},
				FileIDs: []string{},
			},
		},
		{
			Name: "AddedCodeToProject with multiple codes",
			Initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1"},
				FileIDs: []string{},
			},
			Event: newProjectEvent("project-1", project.AddedCodeToProject, &project.AddedCodeToProjectPayload{
				CodeID:    "code-2",
				ProjectID: "project-1",
			}),
			Expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1", "code-2"},
				FileIDs: []string{},
			},
		},
		{
			Name: "RemovedCodeFromProject removes from CodeIDs",
			Initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1", "code-2", "code-3"},
				FileIDs: []string{},
			},
			Event: newProjectEvent("project-1", project.RemovedCodeFromProject, &project.RemovedCodeFromProjectPayload{
				CodeID:    "code-2",
				ProjectID: "project-1",
			}),
			Expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1", "code-3"},
				FileIDs: []string{},
			},
		},
		{
			Name: "RemovedCodeFromProject on non-existent code is safe",
			Initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1"},
				FileIDs: []string{},
			},
			Event: newProjectEvent("project-1", project.RemovedCodeFromProject, &project.RemovedCodeFromProjectPayload{
				CodeID:    "code-999",
				ProjectID: "project-1",
			}),
			Expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1"},
				FileIDs: []string{},
			},
		},
		{
			Name: "Full lifecycle: create, add files, add codes, delete code",
			Initial: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-1", "code-2"},
				FileIDs: []string{"file-1", "file-2"},
			},
			Event: newProjectEvent("project-1", project.RemovedCodeFromProject, &project.RemovedCodeFromProjectPayload{
				CodeID:    "code-1",
				ProjectID: "project-1",
			}),
			Expected: &Project{
				ID:      "project-1",
				Name:    "Research",
				CodeIDs: []string{"code-2"},
				FileIDs: []string{"file-1", "file-2"},
			},
		},
		{
			Name: "UpdatedProject changes name and description",
			Initial: &Project{
				ID:          "project-1",
				Name:        "Old Name",
				Description: "Old description",
				CodeIDs:     []string{"code-1"},
				FileIDs:     []string{"file-1"},
			},
			Event: newProjectEvent("project-1", project.UpdatedProject, &project.UpdatedProjectPayload{
				Name:        "New Name",
				Description: "New description",
			}),
			Expected: &Project{
				ID:          "project-1",
				Name:        "New Name",
				Description: "New description",
				CodeIDs:     []string{"code-1"},
				FileIDs:     []string{"file-1"},
			},
		},
		{
			Name: "UpdatedProject on nil project returns nil",
			Initial: nil,
			Event: newProjectEvent("project-1", project.UpdatedProject, &project.UpdatedProjectPayload{
				Name: "New Name",
			}),
			Expected: nil,
		},
	}

	deletedEntityTests := reducer_helpers.DeletedEntityTests(
		project.EntityName,
		project.DeletedProject,
		createTestProject,
	)

	combinedTests := append(tests, deletedEntityTests...)

	reducer_helpers.RunReducerTests(t, combinedTests, Reducer)
}

func newProjectEvent(aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent[any, any](action, payload, project.EntityName, aggregateID, nil))
}
