package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/lib/text-search/find"
	"hermes-relay/internal/lib/utils"
)

func NewDocumentRouter(store *registry.Store) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		dispatch.LimitOnEntity(document.EntityName,
			dispatch.LimitOnAction(document.CreateDocument,
				registry.ValidateDomain(store, validateCreateDocument,
					dispatch.ToCreateEntityEvent[document.CreateDocumentPayload, document.CreatedDocumentPayload](document.CreateDocument, document.CreatedDocument, createDocumentFromPayload),
				),
			),
			dispatch.ToUpdateEntityEvent[document.UpdateDocumentPayload, document.UpdatedDocumentPayload](document.UpdateDocument, document.UpdatedDocument),
			dispatch.ToEmptyDomainEvent(document.PinDocument, document.PinnedDocument),
			dispatch.ToEmptyDomainEvent(document.UnpinDocument, document.UnpinnedDocument),
			dispatch.ToUpdateEntityEvent[document.AddDocumentTagsPayload, document.AddedDocumentTagsPayload](document.AddDocumentTags, document.AddedDocumentTags),
			dispatch.ToUpdateEntityEvent[document.RemoveDocumentTagsPayload, document.RemovedDocumentTagsPayload](document.RemoveDocumentTags, document.RemovedDocumentTags),
			addAnnotationHandler(store, document.AddDocumentAnnotation, document.AddedAnnotation),
			dispatch.ToUpdateEntityEvent[document.RemoveAnnotationsPayload, document.RemovedAnnotationsPayload](document.RemoveDocumentAnnotations, document.RemovedAnnotations),
			dispatch.ToUpdateEntityEvent[document.UpdateAnnotationPropsPayload, document.UpdatedAnnotationPropsPayload](document.UpdateDocumentAnnotationProps, document.UpdatedAnnotationProps),
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

func addAnnotationHandler(store *registry.Store, commandAction, eventAction commands.Action) dispatch.CommandRouter {
	normalize := func(proj project.Project, payload document.AddAnnotationPayload, msg *commands.AnyMessage) (document.AddAnnotationPayload, error) {
		doc, exists := proj.GetDocument(msg.AggregateID)
		if !exists {
			return payload, utils.FieldError("document_id", "not found")
		}

		docText := document.ExtractDocumentText(doc.Content)
		matchedText, found := find.Find(payload.Annotation.Text, docText)
		if !found {
			return payload, utils.FieldError("annotation.text", "text not found in document")
		}

		payload.Annotation.Text = matchedText
		payload.Annotation.ID = utils.NewAnnotationID()

		return payload, nil
	}

	return dispatch.LimitOnAction(commandAction,
		registry.NormalizeDomain(store, normalize,
			dispatch.ToUpdateEntityEvent[document.AddAnnotationPayload, document.AddedAnnotationPayload](commandAction, eventAction),
		),
	)
}
