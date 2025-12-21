package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/projections/registry"
)

func NewDocumentRouter(registryState *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		dispatch.LimitOnEntity(document.EntityName,
			dispatch.LimitOnAction(document.CreateDocument,
				registry.ValidateDomain(registryState, validateCreateDocument,
					dispatch.ToCreateEntityEvent[document.CreateDocumentPayload, document.CreatedDocumentPayload](document.CreateDocument, document.CreatedDocument, createDocumentFromPayload),
				),
			),
			dispatch.ToUpdateEntityEvent[document.UpdateDocumentPayload, document.UpdatedDocumentPayload](document.UpdateDocument, document.UpdatedDocument),
			dispatch.ToEmptyDomainEvent(document.PinDocument, document.PinnedDocument),
			dispatch.ToEmptyDomainEvent(document.UnpinDocument, document.UnpinnedDocument),
			dispatch.ToUpdateEntityEvent[document.AddDocumentTagsPayload, document.AddedDocumentTagsPayload](document.AddDocumentTags, document.AddedDocumentTags),
			dispatch.ToUpdateEntityEvent[document.RemoveDocumentTagsPayload, document.RemovedDocumentTagsPayload](document.RemoveDocumentTags, document.RemovedDocumentTags),
		),
		dispatch.ToEmptyDomainEvent(document.DeleteDocument, document.DeletedDocument),
	)
}

func createDocumentFromPayload(payload *document.CreateDocumentPayload) document.CreatedDocumentPayload {
	return document.CreatedDocumentPayload{
		ProjectID:   payload.ProjectID,
		Name:        payload.Name,
		Description: payload.Description,
	}
}

func validateCreateDocument(_ project.Project, _ document.CreateDocumentPayload, _ *commands.AnyMessage) error {
	return nil
}
