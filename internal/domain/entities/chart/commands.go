package chart

import (
	"hermes-relay/internal/cqrs/commands"
)

const EntityName commands.AggregateType = "Chart"

const (
	CreateChart commands.Action = "CreateChart"
	UpdateChart commands.Action = "UpdateChart"
	PinChart    commands.Action = "PinChart"
	UnpinChart  commands.Action = "UnpinChart"
	DeleteChart commands.Action = "DeleteChart"
)

type CreateChartPayload = ChartData
type UpdateChartPayload = UpdateChartData
type PinChartPayload = commands.EmptyPayload
type UnpinChartPayload = commands.EmptyPayload
type DeleteChartPayload = commands.EmptyPayload
