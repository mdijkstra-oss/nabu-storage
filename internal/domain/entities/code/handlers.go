package code

import (
	"hermes-relay/internal/cqrs"
)

var Router = cqrs.CombineRouters(
	cqrs.LimitOnEntity(EntityName,
		cqrs.LimitOnType(cqrs.Command,
			cqrs.ToCreateEntityEvent[CreateCodePayload](CreateCode, CreatedCode),
			cqrs.ToUpdateEntityEvent[UpdateCodePayload](UpdateCode, UpdatedCode),
			cqrs.ToEmptyDomainEvent(DeleteCode, DeletedCode),
		),
	),
)
