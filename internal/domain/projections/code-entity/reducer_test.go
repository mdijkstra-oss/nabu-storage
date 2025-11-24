package codeview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
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
			Event: newCodeEvent("code-1", code.CreatedCode, &code.CreatedCodePayload{
				ProjectID: "project-1",
				Slug:      "topic:climate",
				Color:     "green-500",
				Definition: "Climate topics",
			}),
			Expected: buildCode("code-1", code.CodeData{}),
		},
		{
			Name:    "UpdatedCode changes color only",
			Initial: buildCode("code-1", code.CodeData{}),
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color: "emerald-600",
			}),
			Expected: buildCode("code-1", code.CodeData{Color: "emerald-600"}),
		},
		{
			Name:    "UpdatedCode changes definition only",
			Initial: buildCode("code-1", code.CodeData{}),
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Definition: "Renewable energy and sustainability",
			}),
			Expected: buildCode("code-1", code.CodeData{Definition: "Renewable energy and sustainability"}),
		},
		{
			Name:    "UpdatedCode changes both fields",
			Initial: buildCode("code-1", code.CodeData{}),
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color:     "teal-500",
				Definition: "Environmental coverage",
			}),
			Expected: buildCode("code-1", code.CodeData{Color: "teal-500", Definition: "Environmental coverage"}),
		},
		{
			Name:    "UpdatedCode with empty strings preserves current values",
			Initial: buildCode("code-1", code.CodeData{Color: "teal-500", Definition: "Environmental coverage"}),
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
				Color:     "",
				Definition: "",
			}),
			Expected: buildCode("code-1", code.CodeData{Color: "teal-500", Definition: "Environmental coverage"}),
		},
		{
			Name: "UpdatedCode with all empty fields preserves all current values",
			Initial: buildCode("code-1", code.CodeData{
				Slug:              "emotion:anxiety",
				Color:             "amber-600",
				Definition:        "Expressions of worry or fear about future events",
				InclusionCriteria: "Clear statements of worry, fear, or concern",
				ExclusionCriteria: "General stress without specific worry",
				Examples:          []string{"I'm worried about...", "What if something goes wrong?"},
				CounterExamples:   []string{"I'm stressed but managing"},
				Notes:             "Often co-occurs with uncertainty codes",
			}),
			Event: newCodeEvent("code-1", code.UpdatedCode, &code.UpdatedCodePayload{
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
				Color:             "amber-600",
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
			Event:    newCodeEvent("code-1", code.MergedCodes, &code.MergedCodesPayload{SourceID: "code-1", TargetID: "code-2"}),
			Expected: nil,
		},
		{
			Name:    "MergedCodes keeps target code unchanged",
			Initial: buildCode("code-2", code.CodeData{Slug: "topic:temperature", Color: "red-500", Definition: "Temperature topics"}),
			Event:   newCodeEvent("code-2", code.MergedCodes, &code.MergedCodesPayload{SourceID: "code-1", TargetID: "code-2"}),
			Expected: buildCode("code-2", code.CodeData{Slug: "topic:temperature", Color: "red-500", Definition: "Temperature topics"}),
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

func newCodeEvent(aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent(action, payload, code.EntityName, aggregateID, commands.Actor{UserID: "test-user", ActorType: commands.ActorTypeSystem}, (*commands.AnyMessage)(nil)))
}
