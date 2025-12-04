package codeview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
)

var Reducer = projection.WithVersionIncrement(
	projection.WithHealthCheck(
		projection.CombineReducers(
			projection.For(code.CreatedCode, CreatedCodeReducer),
			projection.IfExists(
				projection.For(code.UpdatedCode, projection.UpdatedEntity[Code, code.UpdatedCodePayload]),
				projection.For(code.PinnedCode, projection.PinnedEntity[Code]),
				projection.For(code.UnpinnedCode, projection.UnpinnedEntity[Code]),
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

func MergedCodesReducer(current *Code, message *commands.AnyMessage, payload *code.MergedCodesPayload) *Code {
	if message.AggregateID == payload.SourceID {
		return nil
	}
	return current
}
