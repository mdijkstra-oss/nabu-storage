package project

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
)

// EventHandlers listens to file/code events and publishes project relationship events
var EventHandlers = cqrs.CombineRouters(
	cqrs.LimitOnType(cqrs.DomainEvent,
		cqrs.ForPayload(file.CreatedFile, OnFileCreated),
		cqrs.ForPayload(code.CreatedCode, OnCodeCreated),
		cqrs.ForPayload(code.DeletedCode, OnCodeDeleted),
	),
)

func OnFileCreated(msg *cqrs.AnyMessage, payload file.CreatedFilePayload, publish cqrs.PublishFunc) (*cqrs.AnyMessage, error) {
	return publish(cqrs.ToAny(cqrs.NewDomainEvent(
		AddedFileToProject,
		AddedFileToProjectPayload{
			FileID:    msg.AggregateID,
			ProjectID: payload.ProjectID,
		},
		EntityName,
		payload.ProjectID, // aggregate ID is the project ID
		msg,
	)))
}

func OnCodeCreated(msg *cqrs.AnyMessage, payload code.CreatedCodePayload, publish cqrs.PublishFunc) (*cqrs.AnyMessage, error) {
	return publish(cqrs.ToAny(cqrs.NewDomainEvent(
		AddedCodeToProject,
		AddedCodeToProjectPayload{
			CodeID:    msg.AggregateID,
			ProjectID: payload.ProjectID,
		},
		EntityName,
		payload.ProjectID, // aggregate ID is the project ID
		msg,
	)))
}

func OnCodeDeleted(msg *cqrs.AnyMessage, payload code.DeletedCodePayload, publish cqrs.PublishFunc) (*cqrs.AnyMessage, error) {
	return publish(cqrs.ToAny(cqrs.NewDomainEvent(
		RemovedCodeFromProject,
		RemovedCodeFromProjectPayload{
			CodeID:    msg.AggregateID,
			ProjectID: payload.ProjectID,
		},
		EntityName,
		payload.ProjectID, // aggregate ID is the project ID
		msg,
	)))
}
