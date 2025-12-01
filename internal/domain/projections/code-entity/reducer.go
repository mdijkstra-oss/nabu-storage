package codeview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/lib/utils"
)

var Reducer = projection.WithVersionIncrement(
	projection.WithHealthCheck(
		projection.CombineReducers(
			projection.For(code.CreatedCode, CreatedCodeReducer),
			projection.IfExists(
				projection.For(code.UpdatedCode, UpdatedCodeReducer),
				projection.For(code.DeletedCode, projection.DeletedEntity[Code]),
				projection.For(code.MergedCodes, MergedCodesReducer),
			),
			projection.DeletedProjectReducer[code.Code],
		),
	),
)

func CreatedCodeReducer(_ *Code, message *commands.AnyMessage, payload *code.CreatedCodePayload) *Code {
	return &code.Code{
		ID:       message.AggregateID,
		Healthy:  true,
		CodeData: *payload,
	}
}

func UpdatedCodeReducer(current *Code, _ *commands.AnyMessage, payload *code.UpdatedCodePayload) *Code {
	updated := utils.ApplyPartialUpdate(*current, payload)
	return &updated
}

func MergedCodesReducer(current *Code, message *commands.AnyMessage, payload *code.MergedCodesPayload) *Code {
	if message.AggregateID == payload.SourceID {
		return nil
	}
	return current
}
