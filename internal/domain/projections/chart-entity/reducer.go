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
				projection.For(chart.PinnedChart, PinnedChartReducer),
				projection.For(chart.UnpinnedChart, UnpinnedChartReducer),
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

func PinnedChartReducer(current *Chart, _ *commands.AnyMessage, _ *commands.EmptyPayload) *Chart {
	updated := *current
	updated.Pinned = true
	return &updated
}

func UnpinnedChartReducer(current *Chart, _ *commands.AnyMessage, _ *commands.EmptyPayload) *Chart {
	updated := *current
	updated.Pinned = false
	return &updated
}
