package project

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
)

var Router = dispatch.CombineRouters(
	dispatch.LimitOnEntity(EntityName,
		dispatch.LimitOnType(commands.Command,
			dispatch.ToCreateEntityEvent[CreateProjectPayload](CreateProject, CreatedProject),
		),
	),
)
