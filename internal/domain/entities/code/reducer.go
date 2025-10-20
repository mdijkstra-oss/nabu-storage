package code

import (
	"hermes-relay/internal/cqrs"
)

var Reducer = cqrs.CombineReducers(
	cqrs.For(CreatedCode, CreatedCodeReducer),
	cqrs.For(UpdatedCode, UpdatedCodeReducer),
	cqrs.For(DeletedCode, DeletedCodeReducer),
)

func CreatedCodeReducer(_ *Code, message *cqrs.Message, payload *CreatedCodePayload) *Code {
	return &Code{
		ID:    message.AggregateID,
		Color: payload.Color,
	}
}

func UpdatedCodeReducer(current *Code, _ *cqrs.Message, payload *UpdateCodePayload) *Code {
	current.Color = payload.Color
	return current
}

func DeletedCodeReducer(_ *Code, _ *cqrs.Message, _ any) *Code {
	return nil
}
