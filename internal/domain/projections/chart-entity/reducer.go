package chartview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/chart"
	"hermes-relay/internal/lib/utils"
)

type Chart = chart.Chart

var Reducer = projection.WithVersionIncrement(
	projection.WithHealthCheck(
		projection.CombineReducers(
			projection.For(chart.CreatedChart, CreatedChartReducer),
			projection.IfExists(
				projection.For(chart.UpdatedChart, UpdatedChartReducer),
				projection.For(chart.PinnedChart, projection.PinnedEntity[Chart]),
				projection.For(chart.UnpinnedChart, projection.UnpinnedEntity[Chart]),
				projection.For(chart.DeletedChart, projection.DeletedEntity[Chart]),
			),
			projection.DeletedProjectReducer[chart.Chart],
		),
	),
)

func CreatedChartReducer(_ *Chart, message *commands.AnyMessage, payload *chart.CreatedChartPayload) *Chart {
	return &chart.Chart{
		ID:        message.AggregateID,
		Healthy:   true,
		ChartData: *payload,
	}
}

func UpdatedChartReducer(current *Chart, _ *commands.AnyMessage, payload *chart.UpdatedChartPayload) *Chart {
	updated := utils.ApplyPartialUpdate(*current, payload)
	if payload.Spec != nil {
		updated.Spec = *payload.Spec
	}
	return &updated
}

