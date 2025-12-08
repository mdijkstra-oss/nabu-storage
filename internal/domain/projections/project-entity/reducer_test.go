package projectview

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
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
		{
			Name:    "ChangedPhase updates phase",
			Initial: createTestProject("project-1", "Research", "Description"),
			Event: domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.ChangedPhase, &project.ChangedPhasePayload{
				Phase: project.PhaseCode,
			}),
			Expected: func() *Project {
				p := createTestProject("project-1", "Research", "Description")
				p.Phase = project.PhaseCode
				return p
			}(),
		},
		{
			Name:    "ChangedPhase on nil project returns nil",
			Initial: nil,
			Event: domain_helpers.NewDomainEvent(project.EntityName, "project-1", project.ChangedPhase, &project.ChangedPhasePayload{
				Phase: project.PhaseCode,
			}),
			Expected: nil,
		},
	}

	deletedEntityTests := reducer_helpers.DeletedEntityTests(
		project.EntityName,
		project.DeletedProject,
		func() *Project { return createTestProject("project-1", "Test Project", "Test description") },
	)

	codeChildTests := reducer_helpers.AggregateChildMapTests(reducer_helpers.AggregateChildMapTestConfig[Project, code.Code]{
		CreatedEvent:      domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.CreatedCode, &code.CreatedCodePayload{ProjectID: "project-1", Slug: "test-code", Color: "red", Definition: "A test code"}),
		UpdatedEvent:      domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.UpdatedCode, &code.UpdateCodePayload{Slug: "new-slug"}),
		DeletedEvent:      domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.DeletedCode, nil),
		EntityAfterCreate: code.BuildTestCode("code-1", code.CodeData{Slug: "test-code", Color: "red", Definition: "A test code"}),
		EntityAfterUpdate: code.BuildTestCode("code-1", code.CodeData{Slug: "new-slug", Color: "red", Definition: "A test code"}),
		CreateParent:      createEmptyProject,
		GetMap:            func(p *Project) map[string]code.Code { return p.Codes },
	})

	fileChildTests := reducer_helpers.AggregateChildMapTests(reducer_helpers.AggregateChildMapTestConfig[Project, file.File]{
		CreatedEvent: domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.CreatedFile, &file.CreatedFilePayload{
			FileData: file.FileData{
				ProjectID:   "project-1",
				Name:        "test-file.txt",
				Description: "A test file",
				Type:        file.FileTypeCorpus,
				Locked:      false,
			},
			Content: "",
			Codes:   []file.CodedSection{},
		}),
		UpdatedEvent:      domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.UpdatedFile, &file.UpdatedFilePayload{Name: "new-name.txt", Description: "Updated"}),
		DeletedEvent:      domain_helpers.NewDomainEvent(file.EntityName, "file-1", file.DeletedFile, nil),
		EntityAfterCreate: file.BuildTestFile("file-1", file.FileData{ProjectID: "project-1", Name: "test-file.txt", Description: "A test file"}),
		EntityAfterUpdate: file.BuildTestFile("file-1", file.FileData{ProjectID: "project-1", Name: "new-name.txt", Description: "Updated"}),
		CreateParent:      createEmptyProject,
		GetMap:            func(p *Project) map[string]file.File { return p.Files },
	})

	combinedTests := append(tests, deletedEntityTests...)
	combinedTests = append(combinedTests, codeChildTests...)
	combinedTests = append(combinedTests, fileChildTests...)

	reducer_helpers.RunReducerTests(t, combinedTests, Reducer, clearFileTimestamps)
}

func clearFileTimestamps(proj *Project) *Project {
	if proj == nil {
		return nil
	}
	for id, f := range proj.Files {
		f.Time = th.DefaultTestTime()
		proj.Files[id] = f
	}
	return proj
}
