package codeview

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"hermes-relay/internal/lib/test-helpers/reducer-helpers"
	"testing"
)

func createTestCode(projectID string) *Code {
	c := code.BuildTestCode("code-1", code.CodeData{ProjectID: projectID})
	return &c
}

func buildCode(id string, overrides code.CodeData) *Code {
	c := code.BuildTestCode(id, overrides)
	return &c
}

func TestCodeReducer(t *testing.T) {
	tests := []reducer_helpers.ReducerTestCase[*Code]{
		{
			Name:    "CreatedCode initializes code",
			Initial: nil,
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.CreatedCode, &code.CreatedCodePayload{
				ProjectID:  "project-1",
				Slug:       "topic:climate",
				Color:      "green",
				Definition: "Climate topics",
			}),
			Expected: buildCode("code-1", code.CodeData{}),
		},
		{
			Name:    "UpdatedCode changes color only",
			Initial: buildCode("code-1", code.CodeData{}),
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color: "teal",
			}),
			Expected: buildCode("code-1", code.CodeData{Color: "teal"}),
		},
		{
			Name:    "UpdatedCode changes definition only",
			Initial: buildCode("code-1", code.CodeData{}),
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Definition: "Renewable energy and sustainability",
			}),
			Expected: buildCode("code-1", code.CodeData{Definition: "Renewable energy and sustainability"}),
		},
		{
			Name:    "UpdatedCode changes both fields",
			Initial: buildCode("code-1", code.CodeData{}),
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color:      "cyan",
				Definition: "Environmental coverage",
			}),
			Expected: buildCode("code-1", code.CodeData{Color: "cyan", Definition: "Environmental coverage"}),
		},
		{
			Name:    "UpdatedCode with empty strings preserves current values",
			Initial: buildCode("code-1", code.CodeData{Color: "cyan", Definition: "Environmental coverage"}),
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color:      "",
				Definition: "",
			}),
			Expected: buildCode("code-1", code.CodeData{Color: "cyan", Definition: "Environmental coverage"}),
		},
		{
			Name: "UpdatedCode with all empty fields preserves all current values",
			Initial: buildCode("code-1", code.CodeData{
				Slug:              "emotion:anxiety",
				Color:             "amber",
				Definition:        "Expressions of worry or fear about future events",
				InclusionCriteria: "Clear statements of worry, fear, or concern",
				ExclusionCriteria: "General stress without specific worry",
				Examples:          []string{"I'm worried about...", "What if something goes wrong?"},
				CounterExamples:   []string{"I'm stressed but managing"},
				Notes:             "Often co-occurs with uncertainty codes",
			}),
			Event: domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Slug:              "",
				Color:             "",
				Definition:        "",
				InclusionCriteria: "",
				ExclusionCriteria: "",
				Examples:          nil,
				CounterExamples:   nil,
				Notes:             "",
			}),
			Expected: buildCode("code-1", code.CodeData{
				Slug:              "emotion:anxiety",
				Color:             "amber",
				Definition:        "Expressions of worry or fear about future events",
				InclusionCriteria: "Clear statements of worry, fear, or concern",
				ExclusionCriteria: "General stress without specific worry",
				Examples:          []string{"I'm worried about...", "What if something goes wrong?"},
				CounterExamples:   []string{"I'm stressed but managing"},
				Notes:             "Often co-occurs with uncertainty codes",
			}),
		},
		{
			Name:     "MergedCodes returns nil when processing source code",
			Initial:  buildCode("code-1", code.CodeData{}),
			Event:    domain_helpers.NewDomainEvent(code.EntityName, "code-1", code.MergedCodes, &code.MergedCodesPayload{SourceID: "code-1", TargetID: "code-2"}),
			Expected: nil,
		},
		{
			Name:     "MergedCodes keeps target code unchanged",
			Initial:  buildCode("code-2", code.CodeData{Slug: "topic:temperature", Color: "red", Definition: "Temperature topics"}),
			Event:    domain_helpers.NewDomainEvent(code.EntityName, "code-2", code.MergedCodes, &code.MergedCodesPayload{SourceID: "code-1", TargetID: "code-2"}),
			Expected: buildCode("code-2", code.CodeData{Slug: "topic:temperature", Color: "red", Definition: "Temperature topics"}),
		},
	}

	deletedEntityTests := reducer_helpers.DeletedEntityTests(
		code.EntityName,
		code.DeletedCode,
		func() *Code { return createTestCode("project-1") },
	)
	deletedProjectTests := reducer_helpers.DeletedProjectCascadeTests(createTestCode)

	combinedTests := append(tests, deletedEntityTests...)
	combinedTests = append(combinedTests, deletedProjectTests...)

	reducer_helpers.RunReducerTests(t, combinedTests, Reducer)
}
