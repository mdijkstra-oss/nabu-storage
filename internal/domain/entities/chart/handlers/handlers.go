package handlers

import (
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/chart"
	"hermes-relay/internal/domain/projections/registry"
)

func NewRouter(reg *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.LimitOnEntity(chart.EntityName,
		dispatch.ToCreateEntityEvent[chart.CreateChartPayload, chart.CreatedChartPayload](chart.CreateChart, chart.CreatedChart),
		dispatch.ToUpdateEntityEvent[chart.UpdateChartPayload, chart.UpdatedChartPayload](chart.UpdateChart, chart.UpdatedChart),
		dispatch.ToEmptyDomainEvent(chart.PinChart, chart.PinnedChart),
		dispatch.ToEmptyDomainEvent(chart.UnpinChart, chart.UnpinnedChart),
		dispatch.ToEmptyDomainEvent(chart.DeleteChart, chart.DeletedChart),
	)
}
