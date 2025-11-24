package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/reducer-helpers"
	"testing"
)

func TestRegistryReducer(t *testing.T) {
	tests := []reducer_helpers.ReducerTestCase[*Registry]{
		{
			Name:    "CreatedProject adds project to registry",
			Initial: nil,
			Event: newProjectEvent("project-1", project.CreatedProject, &project.CreatedProjectPayload{
				Name:        "Test Project",
				Description: "A test project",
			}),
			Expected: registryWith(
				projectWith("project-1", "Test Project", "A test project", nil, nil),
			),
		},
		{
			Name:    "CreatedCode adds code to project and updates lookup",
			Initial: registryWith(emptyProject("project-1", "Test Project")),
			Event:   newCodeEvent("code-1", code.CreatedCode, &code.CreatedCodePayload{ProjectID: "project-1", Slug: "test-code", Color: "blue-500", Definition: "Test code"}),
			Expected: registryWith(
				projectWith("project-1", "Test Project", "", map[string]code.Code{"code-1": testCode("code-1", "project-1", "test-code")}, nil),
				withLookup("Code:code-1", "project-1"),
			),
		},
		{
			Name:    "CreatedFile adds file to project and updates lookup",
			Initial: registryWith(emptyProject("project-1", "Test Project")),
			Event:   newFileEvent("file-1", file.CreatedFile, &file.CreatedFilePayload{FileData: file.FileData{ProjectID: "project-1", Name: "test.md", Type: file.FileTypeSource}, Chunks: []file.Chunk{}}),
			Expected: registryWith(
				projectWith("project-1", "Test Project", "", nil, map[string]file.File{"file-1": testFile("file-1", "project-1", "test.md")}),
				withLookup("File:file-1", "project-1"),
			),
		},
		{
			Name: "DeletedCode removes from lookup table",
			Initial: registryWith(
				projectWith("project-1", "Test Project", "", map[string]code.Code{"code-1": testCode("code-1", "project-1", "test-code")}, nil),
				withLookup("Code:code-1", "project-1"),
			),
			Event:    newCodeEvent("code-1", code.DeletedCode, nil),
			Expected: registryWith(emptyProject("project-1", "Test Project")),
		},
		{
			Name:     "DeletedProject removes project from registry",
			Initial:  registryWith(emptyProject("project-1", "Test Project")),
			Event:    newProjectEvent("project-1", project.DeletedProject, nil),
			Expected: registryWith(),
		},
		{
			Name: "UpdatedCode routes to correct project via lookup",
			Initial: registryWith(
				projectWith("project-1", "Test Project", "", map[string]code.Code{"code-1": testCode("code-1", "project-1", "old-slug")}, nil),
				withLookup("Code:code-1", "project-1"),
			),
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdateCodePayload{Slug: "new-slug"}),
			Expected: registryWith(
				projectWith("project-1", "Test Project", "", map[string]code.Code{"code-1": testCode("code-1", "project-1", "new-slug")}, nil),
				withLookup("Code:code-1", "project-1"),
			),
		},
		{
			Name: "Multiple projects remain isolated",
			Initial: registryWith(
				emptyProject("project-1", "Project 1"),
				emptyProject("project-2", "Project 2"),
			),
			Event: newCodeEvent("code-1", code.CreatedCode, &code.CreatedCodePayload{ProjectID: "project-1", Slug: "test-code", Color: "blue-500", Definition: "Test code"}),
			Expected: registryWith(
				projectWith("project-1", "Project 1", "", map[string]code.Code{"code-1": testCode("code-1", "project-1", "test-code")}, nil),
				emptyProject("project-2", "Project 2"),
				withLookup("Code:code-1", "project-1"),
			),
		},
	}

	reducer_helpers.RunReducerTests(t, tests, Reducer, clearFileTimestamps)
}

type registryOpt func(*Registry)

func registryWith(opts ...registryOpt) *Registry {
	reg := &Registry{Projects: make(map[string]project.Project), EntityToProject: make(map[string]string)}
	for _, opt := range opts {
		opt(reg)
	}
	return reg
}

func emptyProject(id, name string) registryOpt {
	return projectWith(id, name, "", nil, nil)
}

func projectWith(id, name, desc string, codes map[string]code.Code, files map[string]file.File) registryOpt {
	return func(r *Registry) {
		if codes == nil {
			codes = make(map[string]code.Code)
		}
		if files == nil {
			files = make(map[string]file.File)
		}
		r.Projects[id] = project.Project{
			ID:          id,
			Healthy:     true,
			ProjectData: project.ProjectData{Name: name, Description: desc},
			Codes:       codes,
			Files:       files,
		}
	}
}

func withLookup(key, projectID string) registryOpt {
	return func(r *Registry) {
		r.EntityToProject[key] = projectID
	}
}

func testCode(id, projectID, slug string) code.Code {
	return code.BuildTestCode(id, code.CodeData{
		ProjectID:  projectID,
		Slug:       slug,
		Color:      "blue-500",
		Definition: "Test code",
	})
}

func testFile(id, projectID, name string) file.File {
	return file.BuildTestFile(id, file.FileData{
		ProjectID: projectID,
		Name:      name,
	})
}

func clearFileTimestamps(reg *Registry) *Registry {
	if reg == nil {
		return nil
	}
	for projID, proj := range reg.Projects {
		for fileID, f := range proj.Files {
			f.Time = th.DefaultTestTime()
			proj.Files[fileID] = f
		}
		reg.Projects[projID] = proj
	}
	return reg
}

func newProjectEvent(aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent[any, any](action, payload, project.EntityName, aggregateID, commands.Actor{UserID: "test-user", ActorType: commands.ActorTypeSystem}, nil))
}

func newCodeEvent(aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent[any, any](action, payload, code.EntityName, aggregateID, commands.Actor{UserID: "test-user", ActorType: commands.ActorTypeSystem}, nil))
}

func newFileEvent(aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent[any, any](action, payload, file.EntityName, aggregateID, commands.Actor{UserID: "test-user", ActorType: commands.ActorTypeSystem}, nil))
}

type resolveProjectIDInput struct {
	Setup func(*RegistryState)
	Event *commands.AnyMessage
}

func TestResolveProjectID(t *testing.T) {
	tests := []struct {
		Name     string
		Input    resolveProjectIDInput
		Expected string
	}{
		{
			Name: "Project aggregate returns aggregate ID directly",
			Input: resolveProjectIDInput{
				Setup: func(rs *RegistryState) {},
				Event: &commands.AnyMessage{
					AggregateType: "Project",
					AggregateID:   "proj-123",
					Payload:       []byte("{}"),
				},
			},
			Expected: "proj-123",
		},
		{
			Name: "File with ProjectID in payload extracts it",
			Input: resolveProjectIDInput{
				Setup: func(rs *RegistryState) {},
				Event: newFileEvent("file-1", file.CreatedFile, &file.CreatedFilePayload{
					FileData: file.FileData{ProjectID: "proj-456", Name: "test.md", Type: file.FileTypeSource},
					Chunks:   []file.Chunk{},
				}),
			},
			Expected: "proj-456",
		},
		{
			Name: "Code with ProjectID in payload extracts it",
			Input: resolveProjectIDInput{
				Setup: func(rs *RegistryState) {},
				Event: newCodeEvent("code-1", code.CreatedCode, &code.CreatedCodePayload{
					ProjectID:  "proj-789",
					Slug:       "test-code",
					Color:      "blue-500",
					Definition: "Test",
				}),
			},
			Expected: "proj-789",
		},
		{
			Name: "File without ProjectID in payload falls back to registry lookup",
			Input: resolveProjectIDInput{
				Setup: func(rs *RegistryState) {
					rs.ApplyEvent(newProjectEvent("proj-100", project.CreatedProject, &project.CreatedProjectPayload{
						Name:        "Test Project",
						Description: "Test",
					}))
					rs.ApplyEvent(newFileEvent("file-99", file.CreatedFile, &file.CreatedFilePayload{
						FileData: file.FileData{ProjectID: "proj-100", Name: "existing.md", Type: file.FileTypeSource},
						Chunks:   []file.Chunk{},
					}))
				},
				Event: &commands.AnyMessage{
					AggregateType: "File",
					AggregateID:   "file-99",
					Payload:       []byte("{}"),
				},
			},
			Expected: "proj-100",
		},
		{
			Name: "Code without ProjectID falls back to registry lookup",
			Input: resolveProjectIDInput{
				Setup: func(rs *RegistryState) {
					rs.ApplyEvent(newProjectEvent("proj-200", project.CreatedProject, &project.CreatedProjectPayload{
						Name: "Test Project",
					}))
					rs.ApplyEvent(newCodeEvent("code-88", code.CreatedCode, &code.CreatedCodePayload{
						ProjectID:  "proj-200",
						Slug:       "test-code",
						Color:      "red-500",
						Definition: "Test",
					}))
				},
				Event: &commands.AnyMessage{
					AggregateType: "Code",
					AggregateID:   "code-88",
					Payload:       []byte("{}"),
				},
			},
			Expected: "proj-200",
		},
		{
			Name: "Unknown entity returns empty string",
			Input: resolveProjectIDInput{
				Setup: func(rs *RegistryState) {},
				Event: &commands.AnyMessage{
					AggregateType: "File",
					AggregateID:   "unknown-file",
					Payload:       []byte("{}"),
				},
			},
			Expected: "",
		},
		{
			Name: "Unknown aggregate type returns empty string",
			Input: resolveProjectIDInput{
				Setup: func(rs *RegistryState) {},
				Event: &commands.AnyMessage{
					AggregateType: "UnknownType",
					AggregateID:   "some-id",
					Payload:       []byte("{}"),
				},
			},
			Expected: "",
		},
	}

	th.RunFunctionTests(t, tests, func(input resolveProjectIDInput) string {
		registryState := NewRegistryState()
		input.Setup(registryState)
		return registryState.ResolveProjectID(input.Event)
	})
}
