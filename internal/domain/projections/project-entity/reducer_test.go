package projectview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	th "hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/reducer-helpers"
	"testing"
)

func buildCode(id string, overrides code.CodeData) code.Code {
	return code.BuildTestCode(id, overrides)
}

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
			Event: newProjectEvent("project-1", project.CreatedProject, &project.CreatedProjectPayload{
				Name:        "Research",
				Description: "Research project",
			}),
			Expected: createTestProject("project-1", "Research", "Research project"),
		},
		{
			Name:    "UpdatedProject changes name and description",
			Initial: createTestProject("project-1", "Old Name", "Old description"),
			Event: newProjectEvent("project-1", project.UpdatedProject, &project.UpdatedProjectPayload{
				Name:        "New Name",
				Description: "New description",
			}),
			Expected: createTestProject("project-1", "New Name", "New description"),
		},
		{
			Name:    "UpdatedProject with empty description preserves existing description",
			Initial: createTestProject("project-1", "COVID-19 Research Study", "A comprehensive qualitative research study examining healthcare workers' experiences during the COVID-19 pandemic"),
			Event: newProjectEvent("project-1", project.UpdatedProject, &project.UpdatedProjectPayload{
				Name:        "COVID-19 HCW Research",
				Description: "",
			}),
			Expected: createTestProject("project-1", "COVID-19 HCW Research", "A comprehensive qualitative research study examining healthcare workers' experiences during the COVID-19 pandemic"),
		},
		{
			Name:    "UpdatedProject on nil project returns nil",
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
		func() *Project { return createTestProject("project-1", "Test Project", "Test description") },
	)

	codeChildTests := reducer_helpers.AggregateChildMapTests(reducer_helpers.AggregateChildMapTestConfig[Project, code.Code]{
		CreatedEvent:      newCodeEvent("code-1", code.CreatedCode, &code.CreatedCodePayload{ProjectID: "project-1", Slug: "test-code", Color: "red-500", Definition: "A test code"}),
		UpdatedEvent:      newCodeEvent("code-1", code.UpdatedCode, &code.UpdateCodePayload{Slug: "new-slug"}),
		DeletedEvent:      newCodeEvent("code-1", code.DeletedCode, nil),
		EntityAfterCreate: buildCode("code-1", code.CodeData{Slug: "test-code", Color: "red-500", Definition: "A test code"}),
		EntityAfterUpdate: buildCode("code-1", code.CodeData{Slug: "new-slug", Color: "red-500", Definition: "A test code"}),
		CreateParent:      createEmptyProject,
		GetMap:            func(p *Project) map[string]code.Code { return p.Codes },
	})

	fileChildTests := reducer_helpers.AggregateChildMapTests(reducer_helpers.AggregateChildMapTestConfig[Project, file.File]{
		CreatedEvent: newFileEvent("file-1", file.CreatedFile, &file.CreatedFilePayload{
			FileData: file.FileData{
				ProjectID:   "project-1",
				Name:        "test-file.txt",
				Description: "A test file",
				Type:        file.FileTypeSource,
				Locked:      false,
			},
			Chunks: []file.Chunk{},
		}),
		UpdatedEvent:      newFileEvent("file-1", file.UpdatedFile, &file.UpdatedFilePayload{Name: "new-name.txt", Description: "Updated"}),
		DeletedEvent:      newFileEvent("file-1", file.DeletedFile, nil),
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

func newProjectEvent(aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent[any, any](action, payload, project.EntityName, aggregateID, nil))
}

func newCodeEvent(aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent[any, any](action, payload, code.EntityName, aggregateID, nil))
}

func newFileEvent(aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent[any, any](action, payload, file.EntityName, aggregateID, nil))
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
