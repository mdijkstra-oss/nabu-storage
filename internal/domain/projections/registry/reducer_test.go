package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
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
				withLookup("doc-1", "project-1"),
			),
		},
		{
			Name: "DeletedDocument removes from lookup table",
			Initial: registryWith(
				projectWith("project-1", "Test Project", "", map[string]document.Document{"doc-1": testDocument("doc-1", "project-1", "test.md")}),
				withLookup("doc-1", "project-1"),
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
				withLookup("doc-1", "project-1"),
			),
			Event: domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.UpdatedDocument, &document.UpdatedDocumentPayload{Name: "new-name.md"}),
			Expected: registryWith(
				projectWith("project-1", "Test Project", "", map[string]document.Document{"doc-1": testDocument("doc-1", "project-1", "new-name.md")}),
				withLookup("doc-1", "project-1"),
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
				withLookup("doc-1", "project-1"),
			),
		},
		{
			Name: "InsertedBlocks routes to correct document via lookup",
			Initial: registryWith(
				projectWith("project-1", "Test Project", "", map[string]document.Document{"doc-1": testDocument("doc-1", "project-1", "test.md")}),
				withLookup("doc-1", "project-1"),
			),
			Event: domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.InsertedBlocks, &document.InsertedBlocksPayload{
				Position: "head:",
				Blocks:   []document.Block{testBlock("block-1", "Hello world")},
			}),
			Expected: registryWith(
				projectWith("project-1", "Test Project", "", map[string]document.Document{
					"doc-1": testDocumentWithContent("doc-1", "project-1", "test.md", []document.Block{testBlock("block-1", "Hello world")}),
				}),
				withLookup("doc-1", "project-1"),
			),
		},
		{
			Name: "DeletedBlocks routes to correct document via lookup",
			Initial: registryWith(
				projectWith("project-1", "Test Project", "", map[string]document.Document{
					"doc-1": testDocumentWithContent("doc-1", "project-1", "test.md", []document.Block{
						testBlock("block-1", "First"),
						testBlock("block-2", "Second"),
					}),
				}),
				withLookup("doc-1", "project-1"),
			),
			Event: domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.DeletedBlocks, &document.DeletedBlocksPayload{
				BlockIDs: []string{"block-1"},
			}),
			Expected: registryWith(
				projectWith("project-1", "Test Project", "", map[string]document.Document{
					"doc-1": testDocumentWithContent("doc-1", "project-1", "test.md", []document.Block{testBlock("block-2", "Second")}),
				}),
				withLookup("doc-1", "project-1"),
			),
		},
		{
			Name: "ReplacedContent routes to correct document via lookup",
			Initial: registryWith(
				projectWith("project-1", "Test Project", "", map[string]document.Document{
					"doc-1": testDocumentWithContent("doc-1", "project-1", "test.md", []document.Block{testBlock("block-1", "Old content")}),
				}),
				withLookup("doc-1", "project-1"),
			),
			Event: domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.ReplacedContent, &document.ReplacedContentPayload{
				Content: []document.Block{testBlock("block-new", "New content")},
			}),
			Expected: registryWith(
				projectWith("project-1", "Test Project", "", map[string]document.Document{
					"doc-1": testDocumentWithContent("doc-1", "project-1", "test.md", []document.Block{testBlock("block-new", "New content")}),
				}),
				withLookup("doc-1", "project-1"),
			),
		},
		{
			Name: "Block events isolated between projects",
			Initial: registryWith(
				projectWith("project-1", "Project 1", "", map[string]document.Document{
					"doc-1": testDocumentWithContent("doc-1", "project-1", "doc1.md", []document.Block{}),
				}),
				projectWith("project-2", "Project 2", "", map[string]document.Document{
					"doc-2": testDocumentWithContent("doc-2", "project-2", "doc2.md", []document.Block{}),
				}),
				withLookup("doc-1", "project-1"),
				withLookup("doc-2", "project-2"),
			),
			Event: domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.InsertedBlocks, &document.InsertedBlocksPayload{
				Position: "head:",
				Blocks:   []document.Block{testBlock("block-1", "Only in doc-1")},
			}),
			Expected: registryWith(
				projectWith("project-1", "Project 1", "", map[string]document.Document{
					"doc-1": testDocumentWithContent("doc-1", "project-1", "doc1.md", []document.Block{testBlock("block-1", "Only in doc-1")}),
				}),
				projectWith("project-2", "Project 2", "", map[string]document.Document{
					"doc-2": testDocumentWithContent("doc-2", "project-2", "doc2.md", []document.Block{}),
				}),
				withLookup("doc-1", "project-1"),
				withLookup("doc-2", "project-2"),
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

func testDocumentWithContent(id, projectID, name string, content []document.Block) document.Document {
	return document.BuildTestDocument(id, document.DocumentData{
		ProjectID: projectID,
		Name:      name,
		Content:   content,
	})
}

func testBlock(id, text string) document.Block {
	return document.Block{
		ID:   id,
		Type: document.BlockTypeParagraph,
		Content: []document.InlineContent{{
			Type: document.InlineTypeText,
			Text: text,
		}},
	}
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
	Setup func(*Store)
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
				Setup: func(store *Store) {},
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
				Setup: func(store *Store) {},
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
				Setup: func(store *Store) {
					projection.Apply(store, domain_helpers.NewDomainEvent(project.EntityName, "proj-100", project.CreatedProject, &project.CreatedProjectPayload{
						Name:        "Test Project",
						Description: "Test",
					}))
					projection.Apply(store, domain_helpers.NewDomainEvent(document.EntityName, "doc-99", document.CreatedDocument, &document.CreatedDocumentPayload{
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
				Setup: func(store *Store) {},
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
				Setup: func(store *Store) {},
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
		store := NewStore()
		input.Setup(store)
		return projection.Read(store, func(r *Registry) string {
			return ResolveProjectID(r, input.Event)
		})
	})
}
