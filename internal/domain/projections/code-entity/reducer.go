package codeview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
)

var Reducer = projection.CombineReducers(
	projection.For(code.CreatedCode, CreatedCodeReducer),
	projection.IfExists(
		projection.For(code.UpdatedCode, UpdatedCodeReducer),
		projection.For(code.DeletedCode, projection.DeletedEntity[Code]),
		projection.For(code.MergedCodes, MergedCodesReducer),
	),
	projection.DeletedProjectReducer[code.Code],
)

func CreatedCodeReducer(_ *Code, message *commands.AnyMessage, payload *code.CreatedCodePayload) *Code {
	return &code.Code{
		ID:        message.AggregateID,
		ProjectID: payload.ProjectID,
		Slug:      payload.Slug,
		Color:     payload.Color,
		Reasoning: payload.Reasoning,
	}
}

func UpdatedCodeReducer(current *Code, _ *commands.AnyMessage, payload *code.UpdatedCodePayload) *Code {
	if payload.Slug != "" {
		current.Slug = payload.Slug
	}
	if payload.Color != "" {
		current.Color = payload.Color
	}
	if payload.Reasoning != "" {
		current.Reasoning = payload.Reasoning
	}
	return current
}

func MergedCodesReducer(current *Code, message *commands.AnyMessage, payload *code.MergedCodesPayload) *Code {
	if message.AggregateID == payload.SourceID {
		return nil
	}
	return current
}
