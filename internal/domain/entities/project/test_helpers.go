package project

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/chart"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"hermes-relay/internal/lib/utils"
)

func BuildTestProject(id string, overrides ProjectData) Project {
	defaults := ProjectData{
		Name:        "Test Project",
		Description: "Test description",
	}
	merged := utils.ApplyPartialUpdate(defaults, overrides)
	return Project{
		ID:          id,
		Healthy:     true,
		Version:     1,
		ProjectData: merged,
		Charts:      make(map[string]chart.Chart),
		Codes:       make(map[string]code.Code),
		Files:       make(map[string]file.File),
	}
}

func CreatedProjectEvent(id string) *commands.AnyMessage {
	return domain_helpers.NewDomainEvent(EntityName, id, CreatedProject, CreatedProjectPayload{
		Name: "Test Project",
	})
}

func BuildTestProjectWithData() Project {
	proj := BuildTestProject("project-1", ProjectData{})
	proj.Codes["code-1"] = code.BuildTestCode("code-1", code.CodeData{Slug: "emotion:anxiety"})
	proj.Codes["code-2"] = code.BuildTestCode("code-2", code.CodeData{Slug: "topic:climate"})

	file1 := file.BuildTestFile("file-1", file.FileData{Name: "Interview A"})
	file1.Content = "First chunk content. Second chunk content."
	file1.Codes = []file.CodedSection{
		{ID: "f1-c1-s1", CodeID: "code-1", Text: "worried about", Confidence: file.ConfidenceHigh, LastActor: commands.Actor{ActorType: commands.ActorTypeHuman}},
		{ID: "f1-c1-s2", CodeID: "code-2", Text: "climate change", Confidence: file.ConfidenceMedium, LastActor: commands.Actor{ActorType: commands.ActorTypeLLM}},
		{ID: "f1-c2-s1", CodeID: "code-1", Text: "feeling anxious", Confidence: file.ConfidenceLow, LastActor: commands.Actor{ActorType: commands.ActorTypeLLM}},
	}
	proj.Files["file-1"] = file1

	file2 := file.BuildTestFile("file-2", file.FileData{Name: "Interview B"})
	file2.Content = "Another chunk"
	file2.Codes = []file.CodedSection{
		{ID: "f2-c1-s1", CodeID: "code-1", Text: "nervous about", Confidence: file.ConfidenceHigh, LastActor: commands.Actor{ActorType: commands.ActorTypeHuman}},
	}
	proj.Files["file-2"] = file2

	return proj
}
