package registry_test

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/registry"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	codeview "hermes-relay/internal/domain/projections/code-entity"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"testing"
)

type entityLookup struct {
	aggregateType commands.AggregateType
	aggregateID   string
}

type entityLookupInput struct {
	events  []*commands.AnyMessage
	lookups []entityLookup
}

func TestEntityLookups(t *testing.T) {
	tests := []struct {
		Name     string
		Input    entityLookupInput
		Expected []string
	}{
		{
			Name: "Created event adds entity to lookup",
			Input: entityLookupInput{
				events: []*commands.AnyMessage{
					domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
					domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, file.CreatedFilePayload{
						CreateFilePayload: file.CreateFilePayload{ProjectID: "project-1", Name: "test.md", Content: "content"},
						Type:              file.FileTypeSource,
						Locked:            true,
					}),
				},
				lookups: []entityLookup{
					{file.EntityName, "file-1"},
				},
			},
			Expected: []string{"project-1"},
		},
		{
			Name: "Deleted event removes entity from lookup",
			Input: entityLookupInput{
				events: []*commands.AnyMessage{
					domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
					domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.CreatedCode, code.CreatedCodePayload{
						ProjectID: "project-1",
						Slug:      "topic:test",
						Color:     "blue",
						Reasoning: "test",
					}),
					domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.DeletedCode, nil),
				},
				lookups: []entityLookup{
					{code.EntityName, "code-1"},
				},
			},
			Expected: []string{""},
		},
		{
			Name: "Non-existent entity returns empty string",
			Input: entityLookupInput{
				events: []*commands.AnyMessage{
					domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
				},
				lookups: []entityLookup{
					{file.EntityName, "nonexistent"},
				},
			},
			Expected: []string{""},
		},
		{
			Name: "Updated events do not affect lookup",
			Input: entityLookupInput{
				events: []*commands.AnyMessage{
					domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
					domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.CreatedCode, code.CreatedCodePayload{
						ProjectID: "project-1",
						Slug:      "topic:test",
						Color:     "blue",
						Reasoning: "test",
					}),
					domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.UpdatedCode, code.UpdatedCodePayload{
						Color:     "red",
						Reasoning: "updated",
					}),
				},
				lookups: []entityLookup{
					{code.EntityName, "code-1"},
				},
			},
			Expected: []string{"project-1"},
		},
		{
			Name: "Multiple entities tracked independently",
			Input: entityLookupInput{
				events: []*commands.AnyMessage{
					domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test 1"}),
					domain_helpers.NewDomainEvent(project.EntityName, "project-2", project.CreatedProject, project.CreatedProjectPayload{Name: "Test 2"}),
					domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, file.CreatedFilePayload{
						CreateFilePayload: file.CreateFilePayload{ProjectID: "project-1", Name: "test1.md", Content: "content"},
						Type:              file.FileTypeSource,
						Locked:            true,
					}),
					domain_helpers.NewDomainEvent(file.EntityName, "file-2", file.CreatedFile, file.CreatedFilePayload{
						CreateFilePayload: file.CreateFilePayload{ProjectID: "project-2", Name: "test2.md", Content: "content"},
						Type:              file.FileTypeSource,
						Locked:            true,
					}),
					domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.CreatedCode, code.CreatedCodePayload{
						ProjectID: "project-1",
						Slug:      "topic:test",
						Color:     "blue",
						Reasoning: "test",
					}),
				},
				lookups: []entityLookup{
					{file.EntityName, "file-1"},
					{file.EntityName, "file-2"},
					{code.EntityName, "code-1"},
				},
			},
			Expected: []string{"project-1", "project-2", "project-1"},
		},
	}

	th.RunFunctionTests(t, tests, func(input entityLookupInput) []string {
		reg := registry.NewProjectViewRegistry(projectview.Reducer, codeview.Reducer, fileview.Reducer)

		for _, event := range input.events {
			projectID := commands.ExtractProjectID(event)
			if projectID == "" {
				projectID = reg.GetProjectIDForEntity(event.AggregateType, event.AggregateID)
			}

			projectView := reg.EnsureProjectExists(event, projectID)
			if projectView != nil {
				projectView.ApplyEventToAllStores(event)
				reg.UpdateEntityLookups(event, projectID)
			}
		}

		results := make([]string, len(input.lookups))
		for i, lookup := range input.lookups {
			results[i] = reg.GetProjectIDForEntity(lookup.aggregateType, lookup.aggregateID)
		}
		return results
	})
}

type ensureProjectInput struct {
	existingProjectID string
	event             *commands.AnyMessage
}

func TestEnsureProjectExists(t *testing.T) {
	tests := []struct {
		Name     string
		Input    ensureProjectInput
		Expected bool
	}{
		{
			Name: "CreatedProject event creates project",
			Input: ensureProjectInput{
				existingProjectID: "",
				event:             domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test"}),
			},
			Expected: true,
		},
		{
			Name: "Non-CreatedProject event returns nil when project does not exist",
			Input: ensureProjectInput{
				existingProjectID: "",
				event:             domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, file.CreatedFilePayload{CreateFilePayload: file.CreateFilePayload{ProjectID: "project-1", Name: "test.md", Content: "content"}, Type: file.FileTypeSource, Locked: true}),
			},
			Expected: false,
		},
		{
			Name: "Returns existing project if already exists",
			Input: ensureProjectInput{
				existingProjectID: "project-1",
				event:             domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, file.CreatedFilePayload{CreateFilePayload: file.CreateFilePayload{ProjectID: "project-1", Name: "test.md", Content: "content"}, Type: file.FileTypeSource, Locked: true}),
			},
			Expected: true,
		},
		{
			Name: "Wrong entity type returns nil",
			Input: ensureProjectInput{
				existingProjectID: "",
				event:             domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.CreatedCode, code.CreatedCodePayload{ProjectID: "project-1", Slug: "topic:test", Color: "blue", Reasoning: "test"}),
			},
			Expected: false,
		},
		{
			Name: "Wrong action returns nil",
			Input: ensureProjectInput{
				existingProjectID: "",
				event:             domain_helpers.NewDomainEvent(project.EntityName, "project-1", "UpdatedProject", nil),
			},
			Expected: false,
		},
	}

	th.RunFunctionTests(t, tests, func(input ensureProjectInput) bool {
		reg := registry.NewProjectViewRegistry(projectview.Reducer, codeview.Reducer, fileview.Reducer)

		if input.existingProjectID != "" {
			createdEvent := domain_helpers.NewDomainEvent(project.EntityName, input.existingProjectID, project.CreatedProject, project.CreatedProjectPayload{Name: "Existing"})
			reg.EnsureProjectExists(createdEvent, input.existingProjectID)
		}

		projectID := commands.ExtractProjectID(input.event)
		result := reg.EnsureProjectExists(input.event, projectID)
		return result != nil
	})
}

func TestGetAllProjectEntities(t *testing.T) {
	tests := []struct {
		Name     string
		Input    []*commands.AnyMessage
		Expected int
	}{
		{
			Name:     "Empty registry returns empty list",
			Input:    []*commands.AnyMessage{},
			Expected: 0,
		},
		{
			Name: "Single project returns one entity",
			Input: []*commands.AnyMessage{
				domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test 1"}),
			},
			Expected: 1,
		},
		{
			Name: "Multiple projects returns all entities",
			Input: []*commands.AnyMessage{
				domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, project.CreatedProjectPayload{Name: "Test 1"}),
				domain_helpers.NewDomainEvent(project.EntityName, "project-2", project.CreatedProject, project.CreatedProjectPayload{Name: "Test 2"}),
				domain_helpers.NewDomainEvent(project.EntityName, "project-3", project.CreatedProject, project.CreatedProjectPayload{Name: "Test 3"}),
			},
			Expected: 3,
		},
	}

	th.RunFunctionTests(t, tests, func(events []*commands.AnyMessage) int {
		reg := registry.NewProjectViewRegistry(projectview.Reducer, codeview.Reducer, fileview.Reducer)

		for _, event := range events {
			projectID := commands.ExtractProjectID(event)
			projectView := reg.EnsureProjectExists(event, projectID)
			if projectView != nil {
				projectView.ApplyEventToAllStores(event)
			}
		}

		return len(reg.GetAllProjectEntities())
	})
}
