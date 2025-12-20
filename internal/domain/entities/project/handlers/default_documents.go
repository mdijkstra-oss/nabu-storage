package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/templates"
	"hermes-relay/internal/lib/utils"
)

func DefaultDocumentsSaga() dispatch.CommandRouter {
	return dispatch.OnEvent(project.CreatedProject, createDefaultDocuments)
}

func createDefaultDocuments(message *commands.AnyMessage, _ project.CreatedProjectPayload) []*commands.AnyMessage {
	return utils.Map(templates.DefaultDocuments(), func(dd templates.DefaultDocument) *commands.AnyMessage {
		return buildCreateDocumentCommand(message.AggregateID, dd, message.Actor)
	})
}

func buildCreateDocumentCommand(projectID string, dd templates.DefaultDocument, actor commands.Actor) *commands.AnyMessage {
	return commands.ToAny(commands.NewCommand[document.CreateDocumentPayload, any](
		document.CreateDocument,
		document.CreateDocumentPayload{
			ProjectID: projectID,
			Name:      dd.Name,
		},
		document.EntityName,
		utils.NewID(),
		actor,
		nil,
	))
}
