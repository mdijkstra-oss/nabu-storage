package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/registry"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
)

func NewRouter(_ *registry.ProjectViewRegistry) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		dispatch.LimitOnEntity(project.EntityName,
			dispatch.ToCreateEntityEvent[project.CreateProjectPayload, project.CreatedProjectPayload](project.CreateProject, project.CreatedProject),
			dispatch.ToUpdateEntityEvent[project.UpdateProjectPayload, project.UpdatedProjectPayload](project.UpdateProject, project.UpdatedProject),
		),
		dispatch.ToEmptyDomainEvent(project.DeleteProject, project.DeletedProject),
		ExternalEventHandlers,
	)
}

var ExternalEventHandlers = dispatch.CombineRouters(
	dispatch.ForPayload(file.CreatedFile, OnFileCreated),
	dispatch.ForPayload(code.CreatedCode, OnCodeCreated),
	dispatch.ForPayload(code.DeletedCode, OnCodeDeleted),
)

func OnFileCreated(msg *commands.AnyMessage, payload file.CreatedFilePayload, publish dispatch.PublishFunc) (*commands.AnyMessage, error) {
	return publish(commands.ToAny(commands.NewDomainEvent(
		project.AddedFileToProject,
		project.AddedFileToProjectPayload{
			FileID:    msg.AggregateID,
			ProjectID: payload.ProjectID,
		},
		project.EntityName,
		payload.ProjectID,
		msg,
	)))
}

func OnCodeCreated(msg *commands.AnyMessage, payload code.CreatedCodePayload, publish dispatch.PublishFunc) (*commands.AnyMessage, error) {
	return publish(commands.ToAny(commands.NewDomainEvent(
		project.AddedCodeToProject,
		project.AddedCodeToProjectPayload{
			CodeID:    msg.AggregateID,
			ProjectID: payload.ProjectID,
		},
		project.EntityName,
		payload.ProjectID,
		msg,
	)))
}

func OnCodeDeleted(msg *commands.AnyMessage, payload code.DeletedCodePayload, publish dispatch.PublishFunc) (*commands.AnyMessage, error) {
	return publish(commands.ToAny(commands.NewDomainEvent(
		project.RemovedCodeFromProject,
		project.RemovedCodeFromProjectPayload{
			CodeID:    msg.AggregateID,
			ProjectID: payload.ProjectID,
		},
		project.EntityName,
		payload.ProjectID,
		msg,
	)))
}
