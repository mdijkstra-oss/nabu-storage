package code

import (
	"context"
	"hermes-relay/internal/cqrs"
)

var Router = cqrs.CombineRouters(
	cqrs.LimitOnEntity(EntityName,
		cqrs.LimitOnType(cqrs.Command,
			cqrs.ForPayload[CreateCodePayload](CreateCode, CreateCodeHandler),
			cqrs.ToUpdateEntityEvent[UpdateCodePayload](UpdateCode, UpdatedCode),
			cqrs.ToEmptyDomainEvent(DeleteCode, DeletedCode),
		),
	),
)

func CreateCodeHandler(ctx context.Context, message *cqrs.AnyMessage, payload CreateCodePayload, publisher cqrs.PublishFunc) (*cqrs.AnyMessage, error) {
	return cqrs.ToDomainEvent(message, CreatedCode), nil
}
