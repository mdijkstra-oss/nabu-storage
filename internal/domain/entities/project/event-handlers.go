package project

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
)

var EventHandlers = dispatch.CombineRouters(
	dispatch.ForPayload(file.CreatedFile, OnFileCreated),
	dispatch.ForPayload(code.CreatedCode, OnCodeCreated),
	dispatch.ForPayload(code.DeletedCode, OnCodeDeleted),
)

func OnFileCreated(msg *commands.AnyMessage, payload file.CreatedFilePayload, publish dispatch.PublishFunc) (*commands.AnyMessage, error) {
	return publish(commands.ToAny(commands.NewDomainEvent(
		AddedFileToProject,
		AddedFileToProjectPayload{
			FileID:    msg.AggregateID,
			ProjectID: payload.ProjectID,
		},
		EntityName,
		payload.ProjectID,
		msg,
	)))
}

func OnCodeCreated(msg *commands.AnyMessage, payload code.CreatedCodePayload, publish dispatch.PublishFunc) (*commands.AnyMessage, error) {
	return publish(commands.ToAny(commands.NewDomainEvent(
		AddedCodeToProject,
		AddedCodeToProjectPayload{
			CodeID:    msg.AggregateID,
			ProjectID: payload.ProjectID,
		},
		EntityName,
		payload.ProjectID,
		msg,
	)))
}

func OnCodeDeleted(msg *commands.AnyMessage, payload code.DeletedCodePayload, publish dispatch.PublishFunc) (*commands.AnyMessage, error) {
	return publish(commands.ToAny(commands.NewDomainEvent(
		RemovedCodeFromProject,
		RemovedCodeFromProjectPayload{
			CodeID:    msg.AggregateID,
			ProjectID: payload.ProjectID,
		},
		EntityName,
		payload.ProjectID,
		msg,
	)))
}
