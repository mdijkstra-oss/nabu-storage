package chart

import (
	"hermes-relay/internal/cqrs/commands"
)

const (
	CreatedChart commands.Action = "CreatedChart"
	UpdatedChart commands.Action = "UpdatedChart"
	PinnedChart  commands.Action = "PinnedChart"
	UnpinnedChart commands.Action = "UnpinnedChart"
	DeletedChart commands.Action = "DeletedChart"
)

type CreatedChartPayload = ChartData
type UpdatedChartPayload = UpdateChartData
type PinnedChartPayload = commands.EmptyPayload
type UnpinnedChartPayload = commands.EmptyPayload
type DeletedChartPayload = commands.EmptyPayload
