package projectview

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func newProjectEvent(aggregateID string, action cqrs.Action, payload any) *cqrs.AnyMessage {
	return cqrs.ToAny(cqrs.NewDomainEvent(action, payload, project.EntityName, aggregateID, (*cqrs.AnyMessage)(nil)))
}

func newFileEvent(aggregateID string, action cqrs.Action, payload any) *cqrs.AnyMessage {
	return cqrs.ToAny(cqrs.NewDomainEvent(action, payload, file.EntityName, aggregateID, (*cqrs.AnyMessage)(nil)))
}

func newCodeEvent(aggregateID string, action cqrs.Action, payload any) *cqrs.AnyMessage {
	return cqrs.ToAny(cqrs.NewDomainEvent(action, payload, code.EntityName, aggregateID, (*cqrs.AnyMessage)(nil)))
}

func TestProjectReducer_CreatedProject(t *testing.T) {
	var state *Project

	state = Reducer(state, newProjectEvent("project-1", project.CreatedProject, &project.CreatedProjectPayload{
		Name: "Research Project",
	}))

	th.AssertEqual(t, state, &Project{
		ID:      "project-1",
		Name:    "Research Project",
		CodeIDs: []string{},
		FileIDs: []string{},
	}, "After creating project")
}

func TestProjectReducer_FileCreated(t *testing.T) {
	state := &Project{
		ID:      "project-1",
		Name:    "Research Project",
		CodeIDs: []string{},
		FileIDs: []string{},
	}

	// Add file to project
	state = Reducer(state, newFileEvent("file-1", file.CreatedFile, &file.CreatedFilePayload{
		BaseFile: file.BaseFile{
			ProjectID: "project-1",
			Name:      "document.txt",
			Attributes: file.Attributes{
				Title: "Document",
			},
		},
		Content: "Content",
	}))

	th.AssertEqual(t, state.FileIDs, []string{"file-1"}, "After adding file")

	// Add another file
	state = Reducer(state, newFileEvent("file-2", file.CreatedFile, &file.CreatedFilePayload{
		BaseFile: file.BaseFile{
			ProjectID: "project-1",
			Name:      "notes.txt",
			Attributes: file.Attributes{
				Title: "Notes",
			},
		},
		Content: "Notes content",
	}))

	th.AssertEqual(t, state.FileIDs, []string{"file-1", "file-2"}, "After adding second file")

	// File for different project should not be added
	state = Reducer(state, newFileEvent("file-3", file.CreatedFile, &file.CreatedFilePayload{
		BaseFile: file.BaseFile{
			ProjectID: "project-2",
			Name:      "other.txt",
			Attributes: file.Attributes{
				Title: "Other",
			},
		},
		Content: "Other content",
	}))

	th.AssertEqual(t, state.FileIDs, []string{"file-1", "file-2"}, "After file from different project (should not add)")
}

func TestProjectReducer_CodeCreated(t *testing.T) {
	state := &Project{
		ID:      "project-1",
		Name:    "Research Project",
		CodeIDs: []string{},
		FileIDs: []string{},
	}

	// Add code to project
	state = Reducer(state, newCodeEvent("code-1", code.CreatedCode, &code.CreatedCodePayload{
		ProjectID: "project-1",
		Slug:      "topic:climate",
		Color:     "green-500",
		Reasoning: "Climate topics",
	}))

	th.AssertEqual(t, state.CodeIDs, []string{"code-1"}, "After adding code")

	// Add another code
	state = Reducer(state, newCodeEvent("code-2", code.CreatedCode, &code.CreatedCodePayload{
		ProjectID: "project-1",
		Slug:      "topic:policy",
		Color:     "blue-500",
		Reasoning: "Policy topics",
	}))

	th.AssertEqual(t, state.CodeIDs, []string{"code-1", "code-2"}, "After adding second code")

	// Code for different project should not be added
	state = Reducer(state, newCodeEvent("code-3", code.CreatedCode, &code.CreatedCodePayload{
		ProjectID: "project-2",
		Slug:      "topic:other",
		Color:     "red-500",
		Reasoning: "Other topics",
	}))

	th.AssertEqual(t, state.CodeIDs, []string{"code-1", "code-2"}, "After code from different project (should not add)")
}

func TestProjectReducer_CodeDeleted(t *testing.T) {
	state := &Project{
		ID:      "project-1",
		Name:    "Research Project",
		CodeIDs: []string{"code-1", "code-2", "code-3"},
		FileIDs: []string{},
	}

	// Delete code
	state = Reducer(state, newCodeEvent("code-2", code.DeletedCode, nil))

	th.AssertEqual(t, state.CodeIDs, []string{"code-1", "code-3"}, "After deleting code-2")

	// Delete non-existent code (should not error)
	state = Reducer(state, newCodeEvent("code-999", code.DeletedCode, nil))

	th.AssertEqual(t, state.CodeIDs, []string{"code-1", "code-3"}, "After deleting non-existent code")
}

func TestProjectReducer_FullLifecycle(t *testing.T) {
	var state *Project

	// Create project
	state = Reducer(state, newProjectEvent("project-1", project.CreatedProject, &project.CreatedProjectPayload{
		Name: "Research Project",
	}))

	// Add files
	state = Reducer(state, newFileEvent("file-1", file.CreatedFile, &file.CreatedFilePayload{
		BaseFile: file.BaseFile{
			ProjectID: "project-1",
			Name:      "doc1.txt",
			Attributes: file.Attributes{
				Title: "Document 1",
			},
		},
		Content: "Content 1",
	}))

	state = Reducer(state, newFileEvent("file-2", file.CreatedFile, &file.CreatedFilePayload{
		BaseFile: file.BaseFile{
			ProjectID: "project-1",
			Name:      "doc2.txt",
			Attributes: file.Attributes{
				Title: "Document 2",
			},
		},
		Content: "Content 2",
	}))

	// Add codes
	state = Reducer(state, newCodeEvent("code-1", code.CreatedCode, &code.CreatedCodePayload{
		ProjectID: "project-1",
		Slug:      "topic:climate",
		Color:     "green-500",
		Reasoning: "Climate topics",
	}))

	state = Reducer(state, newCodeEvent("code-2", code.CreatedCode, &code.CreatedCodePayload{
		ProjectID: "project-1",
		Slug:      "topic:policy",
		Color:     "blue-500",
		Reasoning: "Policy topics",
	}))

	th.AssertEqual(t, state, &Project{
		ID:      "project-1",
		Name:    "Research Project",
		FileIDs: []string{"file-1", "file-2"},
		CodeIDs: []string{"code-1", "code-2"},
	}, "After full lifecycle")

	// Delete a code
	state = Reducer(state, newCodeEvent("code-1", code.DeletedCode, nil))

	th.AssertEqual(t, state, &Project{
		ID:      "project-1",
		Name:    "Research Project",
		FileIDs: []string{"file-1", "file-2"},
		CodeIDs: []string{"code-2"},
	}, "After deleting code-1")
}
