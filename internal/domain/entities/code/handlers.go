package code

import (
	"context"
	"hermes-relay/internal/cqrs"
)

var Router = cqrs.CombineRouters(
	cqrs.LimitOnEntity(Code{},
		cqrs.LimitOnType(cqrs.Command,
			cqrs.ForPayload[CreateCodePayload](CreateCode, CreateCodeHandler),
			cqrs.ToUpdateEvent[UpdateCodePayload](UpdateCode, UpdatedCode),
			cqrs.ToEmptyDomainEvent(DeleteCode, DeletedCode),
		),
	),
)

func CreateCodeHandler(ctx context.Context, message *cqrs.Message, payload CreateCodePayload, publisher cqrs.PublishFunc) (*cqrs.Message, error) {
	return cqrs.ToDomainEvent(message, CreatedCode), nil
}
