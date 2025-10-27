package project

import (
	"hermes-relay/internal/cqrs"
)

// Router is the combined command router for Project entity
var Router = cqrs.CombineRouters(
	cqrs.LimitOnEntity(EntityName,
		cqrs.LimitOnType(cqrs.Command,
			cqrs.ToCreateEntityEvent[CreateProjectPayload](CreateProject, CreatedProject),
		),
	),
)
