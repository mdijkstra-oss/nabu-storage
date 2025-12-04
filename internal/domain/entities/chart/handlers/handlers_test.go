package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/chart"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	rh "hermes-relay/internal/lib/test-helpers/router-helpers"
	"hermes-relay/internal/lib/utils"
	"testing"
)

var (
	testProjectID = utils.NewID()
	testChartID   = utils.NewID()
)

var cmds = []*commands.AnyMessage{
	commands.ToAny(commands.NewDomainEvent[any, any](
		"CreatedProject",
		map[string]any{"name": "Test Project"},
		"Project",
		testProjectID,
		domain_helpers.TestActor(),
		nil,
	)),
	commands.ToAny(commands.NewDomainEvent[chart.CreatedChartPayload, any](
		chart.CreatedChart,
		chart.CreatedChartPayload{
			ProjectID:   testProjectID,
			Name:        "Test Chart",
			Description: "A test chart",
			Query:       "SELECT * FROM codes",
			Spec: chart.ChartSpec{
				Type: chart.ChartTypeBar,
				X:    "code_slug",
				Y:    "count",
			},
			Pinned: false,
		},
		chart.EntityName,
		testChartID,
		domain_helpers.TestActor(),
		nil,
	)),
}

func TestChartRouter(t *testing.T) {
	tests := []rh.RouterTestCase{
		{
			Name: "CreateChart with valid payload",
			Input: commands.ToAny(commands.NewCommand[chart.CreateChartPayload, any](chart.CreateChart, chart.CreateChartPayload{
				ProjectID:   testProjectID,
				Name:        "Code Frequency",
				Description: "Shows code frequency",
				Query:       "SELECT code_slug, COUNT(*) as count FROM coded_sections GROUP BY code_slug",
				Spec: chart.ChartSpec{
					Type: chart.ChartTypeBar,
					X:    "code_slug",
					Y:    "count",
				},
			}, chart.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[chart.CreatedChartPayload, any](chart.CreatedChart, chart.CreatedChartPayload{
				ProjectID:   testProjectID,
				Name:        "Code Frequency",
				Description: "Shows code frequency",
				Query:       "SELECT code_slug, COUNT(*) as count FROM coded_sections GROUP BY code_slug",
				Spec: chart.ChartSpec{
					Type: chart.ChartTypeBar,
					X:    "code_slug",
					Y:    "count",
				},
			}, chart.EntityName, "", domain_helpers.TestActor(), nil)),
		},
		{
			Name: "UpdateChart with valid payload",
			Input: commands.ToAny(commands.NewCommand[chart.UpdateChartPayload, any](chart.UpdateChart, chart.UpdateChartPayload{
				Name:        "Updated Chart Name",
				Description: "Updated description",
			}, chart.EntityName, testChartID, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[chart.UpdatedChartPayload, any](chart.UpdatedChart, chart.UpdatedChartPayload{
				Name:        "Updated Chart Name",
				Description: "Updated description",
			}, chart.EntityName, testChartID, domain_helpers.TestActor(), nil)),
		},
		{
			Name:        "PinChart",
			Input:       commands.ToAny(commands.NewCommand[chart.PinChartPayload, any](chart.PinChart, chart.PinChartPayload{}, chart.EntityName, testChartID, domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[any, any](chart.PinnedChart, nil, chart.EntityName, testChartID, domain_helpers.TestActor(), nil)),
		},
		{
			Name:        "UnpinChart",
			Input:       commands.ToAny(commands.NewCommand[chart.UnpinChartPayload, any](chart.UnpinChart, chart.UnpinChartPayload{}, chart.EntityName, testChartID, domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[any, any](chart.UnpinnedChart, nil, chart.EntityName, testChartID, domain_helpers.TestActor(), nil)),
		},
		{
			Name:        "DeleteChart",
			Input:       commands.ToAny(commands.NewCommand[chart.DeleteChartPayload, any](chart.DeleteChart, chart.DeleteChartPayload{}, chart.EntityName, testChartID, domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[any, any](chart.DeletedChart, nil, chart.EntityName, testChartID, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "CreateChart with missing required fields",
			Input: commands.ToAny(commands.NewCommand[chart.CreateChartPayload, any](chart.CreateChart, chart.CreateChartPayload{
				ProjectID: testProjectID,
			}, chart.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Name is required, Query is required, Type is required, X is required, Y is required",
		},
		{
			Name: "CreateChart with invalid chart type",
			Input: commands.ToAny(commands.NewCommand[chart.CreateChartPayload, any](chart.CreateChart, chart.CreateChartPayload{
				ProjectID: testProjectID,
				Name:      "Invalid Type Chart",
				Query:     "SELECT * FROM test",
				Spec: chart.ChartSpec{
					Type: "invalid",
					X:    "x",
					Y:    "y",
				},
			}, chart.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Type failed validation (oneof)",
		},
		{
			Name: "Wrong entity type returns nil",
			Input: commands.ToAny(commands.NewCommand[chart.CreateChartPayload, any](chart.CreateChart, chart.CreateChartPayload{
				ProjectID: testProjectID,
				Name:      "Test",
				Query:     "SELECT 1",
				Spec: chart.ChartSpec{
					Type: chart.ChartTypeBar,
					X:    "x",
					Y:    "y",
				},
			}, "DifferentEntity", "", domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
	}

	rh.RunRouterTests(t, cmds, tests, NewRouter)
}
