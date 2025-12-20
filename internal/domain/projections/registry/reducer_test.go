package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"hermes-relay/internal/lib/test-helpers/reducer-helpers"
	"testing"
)

func TestRegistryReducer(t *testing.T) {
	tests := []reducer_helpers.ReducerTestCase[*Registry]{
		{
			Name:    "CreatedProject adds project to registry",
			Initial: nil,
			Event: domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, &project.CreatedProjectPayload{
				Name:        "Test Project",
				Description: "A test project",
			}),
			Expected: registryWith(
				projectWith("project-1", "Test Project", "A test project", nil),
			),
		},
		{
			Name:    "CreatedDocument adds document to project and updates lookup",
			Initial: registryWith(emptyProject("project-1", "Test Project")),
			Event:   domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.CreatedDocument, &document.CreatedDocumentPayload{ProjectID: "project-1", Name: "test.md"}),
			Expected: registryWith(
				projectWith("project-1", "Test Project", "", map[string]document.Document{"doc-1": testDocument("doc-1", "project-1", "test.md")}),
				withLookup("Document:doc-1", "project-1"),
			),
		},
		{
			Name: "DeletedDocument removes from lookup table",
			Initial: registryWith(
				projectWith("project-1", "Test Project", "", map[string]document.Document{"doc-1": testDocument("doc-1", "project-1", "test.md")}),
				withLookup("Document:doc-1", "project-1"),
			),
			Event:    domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.DeletedDocument, nil),
			Expected: registryWith(emptyProject("project-1", "Test Project")),
		},
		{
			Name:     "DeletedProject removes project from registry",
			Initial:  registryWith(emptyProject("project-1", "Test Project")),
			Event:    domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.DeletedProject, nil),
			Expected: registryWith(),
		},
		{
			Name: "UpdatedDocument routes to correct project via lookup",
			Initial: registryWith(
				projectWith("project-1", "Test Project", "", map[string]document.Document{"doc-1": testDocument("doc-1", "project-1", "old-name.md")}),
				withLookup("Document:doc-1", "project-1"),
			),
			Event: domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.UpdatedDocument, &document.UpdatedDocumentPayload{Name: "new-name.md"}),
			Expected: registryWith(
				projectWith("project-1", "Test Project", "", map[string]document.Document{"doc-1": testDocument("doc-1", "project-1", "new-name.md")}),
				withLookup("Document:doc-1", "project-1"),
			),
		},
		{
			Name: "Multiple projects remain isolated",
			Initial: registryWith(
				emptyProject("project-1", "Project 1"),
				emptyProject("project-2", "Project 2"),
			),
			Event: domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.CreatedDocument, &document.CreatedDocumentPayload{ProjectID: "project-1", Name: "test.md"}),
			Expected: registryWith(
				projectWith("project-1", "Project 1", "", map[string]document.Document{"doc-1": testDocument("doc-1", "project-1", "test.md")}),
				emptyProject("project-2", "Project 2"),
				withLookup("Document:doc-1", "project-1"),
			),
		},
	}

	reducer_helpers.RunReducerTests(t, tests, Reducer, normalizeRegistry)
}

type registryOpt func(*Registry)

func registryWith(opts ...registryOpt) *Registry {
	reg := &Registry{
		Projects:        make(map[string]project.Project),
		EntityToProject: make(map[string]string),
		Events:          make(map[string][]commands.AnyMessage),
	}
	for _, opt := range opts {
		opt(reg)
	}
	return reg
}

func emptyProject(id, name string) registryOpt {
	return projectWith(id, name, "", nil)
}

func projectWith(id, name, desc string, documents map[string]document.Document) registryOpt {
	return func(r *Registry) {
		if documents == nil {
			documents = make(map[string]document.Document)
		}
		r.Projects[id] = project.Project{
			ID:          id,
			Healthy:     true,
			Version:     1,
			ProjectData: project.ProjectData{Name: name, Description: desc},
			Documents:   documents,
		}
	}
}

func withLookup(key, projectID string) registryOpt {
	return func(r *Registry) {
		r.EntityToProject[key] = projectID
	}
}

func testDocument(id, projectID, name string) document.Document {
	return document.BuildTestDocument(id, document.DocumentData{
		ProjectID: projectID,
		Name:      name,
	})
}

func normalizeRegistry(reg *Registry) *Registry {
	if reg == nil {
		return nil
	}
	for projID, proj := range reg.Projects {
		for docID, d := range proj.Documents {
			d.Time = th.DefaultTestTime()
			proj.Documents[docID] = d
		}
		reg.Projects[projID] = proj
	}
	reg.Events = make(map[string][]commands.AnyMessage)
	return reg
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
			Name: "Document with ProjectID in payload extracts it",
			Input: resolveProjectIDInput{
				Setup: func(rs *RegistryState) {},
				Event: domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.CreatedDocument, &document.CreatedDocumentPayload{
					ProjectID: "proj-456",
					Name:      "test.md",
				}),
			},
			Expected: "proj-456",
		},
		{
			Name: "Document without ProjectID in payload falls back to registry lookup",
			Input: resolveProjectIDInput{
				Setup: func(rs *RegistryState) {
					rs.ApplyEvent(domain_helpers.NewDomainEvent(project.EntityName, "proj-100", project.CreatedProject, &project.CreatedProjectPayload{
						Name:        "Test Project",
						Description: "Test",
					}))
					rs.ApplyEvent(domain_helpers.NewDomainEvent(document.EntityName, "doc-99", document.CreatedDocument, &document.CreatedDocumentPayload{
						ProjectID: "proj-100",
						Name:      "existing.md",
					}))
				},
				Event: &commands.AnyMessage{
					AggregateType: "Document",
					AggregateID:   "doc-99",
					Payload:       []byte("{}"),
				},
			},
			Expected: "proj-100",
		},
		{
			Name: "Unknown entity returns empty string",
			Input: resolveProjectIDInput{
				Setup: func(rs *RegistryState) {},
				Event: &commands.AnyMessage{
					AggregateType: "Document",
					AggregateID:   "unknown-doc",
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
