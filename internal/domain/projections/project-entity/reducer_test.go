package projectview

import (
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"hermes-relay/internal/lib/test-helpers/reducer-helpers"
	"testing"
)

func createTestProject(id, name, description string) *Project {
	p := project.BuildTestProject(id, project.ProjectData{
		Name:        name,
		Description: description,
	})
	return &p
}

func createEmptyProject() *Project {
	return createTestProject("project-1", "Test", "Test")
}

func TestProjectReducer(t *testing.T) {
	tests := []reducer_helpers.ReducerTestCase[*Project]{
		{
			Name:    "CreatedProject initializes empty maps",
			Initial: nil,
			Event: domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.CreatedProject, &project.CreatedProjectPayload{
				Name:        "Research",
				Description: "Research project",
			}),
			Expected: createTestProject("project-1", "Research", "Research project"),
		},
		{
			Name:    "UpdatedProject changes name and description",
			Initial: createTestProject("project-1", "Old Name", "Old description"),
			Event: domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.UpdatedProject, &project.UpdatedProjectPayload{
				Name:        "New Name",
				Description: "New description",
			}),
			Expected: createTestProject("project-1", "New Name", "New description"),
		},
		{
			Name:    "UpdatedProject with empty description preserves existing description",
			Initial: createTestProject("project-1", "COVID-19 Research Study", "A comprehensive qualitative research study examining healthcare workers' experiences during the COVID-19 pandemic"),
			Event: domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.UpdatedProject, &project.UpdatedProjectPayload{
				Name:        "COVID-19 HCW Research",
				Description: "",
			}),
			Expected: createTestProject("project-1", "COVID-19 HCW Research", "A comprehensive qualitative research study examining healthcare workers' experiences during the COVID-19 pandemic"),
		},
		{
			Name:    "UpdatedProject on nil project returns nil",
			Initial: nil,
			Event: domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.UpdatedProject, &project.UpdatedProjectPayload{
				Name: "New Name",
			}),
			Expected: nil,
		},
	}

	deletedEntityTests := reducer_helpers.DeletedEntityTests(
		project.EntityName,
		project.DeletedProject,
		func() *Project { return createTestProject("project-1", "Test Project", "Test description") },
	)

	documentChildTests := reducer_helpers.AggregateChildMapTests(reducer_helpers.AggregateChildMapTestConfig[Project, document.Document]{
		CreatedEvent: domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.CreatedDocument, &document.CreatedDocumentPayload{
			ProjectID: "project-1",
			Name:      "test-doc.txt",
		}),
		UpdatedEvent:      domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.UpdatedDocument, &document.UpdatedDocumentPayload{Name: "new-name.txt", Description: "Updated"}),
		DeletedEvent:      domain_helpers.NewDomainEvent(document.EntityName, "doc-1", document.DeletedDocument, nil),
		EntityAfterCreate: document.BuildTestDocument("doc-1", document.DocumentData{ProjectID: "project-1", Name: "test-doc.txt"}),
		EntityAfterUpdate: document.BuildTestDocument("doc-1", document.DocumentData{ProjectID: "project-1", Name: "new-name.txt", Description: "Updated"}),
		CreateParent:      createEmptyProject,
		GetMap:            func(p *Project) map[string]document.Document { return p.Documents },
	})

	combinedTests := append(tests, deletedEntityTests...)
	combinedTests = append(combinedTests, documentChildTests...)

	reducer_helpers.RunReducerTests(t, combinedTests, Reducer, normalizeProject)
}

func normalizeProject(proj *Project) *Project {
	if proj == nil {
		return nil
	}
	newDocuments := make(map[string]document.Document, len(proj.Documents))
	for id, d := range proj.Documents {
		d.Time = th.DefaultTestTime()
		newDocuments[id] = d
	}
	updated := *proj
	updated.Documents = newDocuments
	return &updated
}
