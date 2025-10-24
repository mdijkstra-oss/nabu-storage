package codeview

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/code"
)

var Reducer = cqrs.CombineReducers(
	cqrs.For(code.CreatedCode, CreatedCodeReducer),
	cqrs.For(code.UpdatedCode, UpdatedCodeReducer),
	cqrs.For(code.DeletedCode, DeletedCodeReducer),
)

func CreatedCodeReducer(_ *Code, message *cqrs.AnyMessage, payload *code.CreatedCodePayload) *Code {
	return &code.Code{
		ID:        message.AggregateID,
		Slug:      payload.Slug,
		Color:     payload.Color,
		Reasoning: payload.Reasoning,
	}
}

func UpdatedCodeReducer(current *Code, _ *cqrs.AnyMessage, payload *code.UpdatedCodePayload) *Code {
	if payload.Color != "" {
		current.Color = payload.Color
	}
	if payload.Reasoning != "" {
		current.Reasoning = payload.Reasoning
	}
	return current
}

func DeletedCodeReducer(_ *Code, _ *cqrs.AnyMessage, _ any) *Code {
	return nil
}
