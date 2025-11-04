package project

import (
	"hermes-relay/internal/cqrs"
)

var Router = cqrs.CombineRouters(
	cqrs.LimitOnEntity(EntityName,
		cqrs.LimitOnType(cqrs.Command,
			cqrs.ToCreateEntityEvent[CreateProjectPayload](CreateProject, CreatedProject),
		),
	),
)
