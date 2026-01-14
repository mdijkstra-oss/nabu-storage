package documentview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/document"
	"time"
)

var DocumentReducer = projection.CombineReducers(
	projection.For(document.CreatedDocument, createdDocumentReducer),
	projection.IfExists(
		projection.For(document.UpdatedDocument, projection.UpdatedEntity[Document, document.UpdatedDocumentPayload]),
		projection.For(document.PinnedDocument, projection.PinnedEntity[Document]),
		projection.For(document.UnpinnedDocument, projection.UnpinnedEntity[Document]),
		projection.For(document.DeletedDocument, projection.DeletedEntity[Document]),
		projection.For(document.AddedDocumentTags, addedDocumentTagsReducer),
		projection.For(document.RemovedDocumentTags, removedDocumentTagsReducer),
		projection.For(document.AddedAnnotation, addedAnnotationReducer),
		projection.For(document.RemovedAnnotations, removedAnnotationsReducer),
		projection.For(document.UpdatedAnnotationProps, updatedAnnotationPropsReducer),
	),
)

func createdDocumentReducer(_ *Document, message *commands.AnyMessage, payload *document.CreatedDocumentPayload) *Document {
	now := time.Now()
	return &Document{
		ID:      message.AggregateID,
		Healthy: true,
		DocumentData: document.DocumentData{
			ProjectID:   payload.ProjectID,
			Name:        payload.Name,
			Description: payload.Description,
			Time:        now,
			UpdatedAt:   now,
			Content:     []document.Block{},
			Tags:        []string{},
			Annotations: []document.Annotation{},
		},
	}
}

func addedDocumentTagsReducer(current *Document, _ *commands.AnyMessage, payload *document.AddedDocumentTagsPayload) *Document {
	current.Tags = document.AddTags(current.Tags, payload.Tags)
	return current
}

func removedDocumentTagsReducer(current *Document, _ *commands.AnyMessage, payload *document.RemovedDocumentTagsPayload) *Document {
	current.Tags = document.RemoveTags(current.Tags, payload.Tags)
	return current
}

func addedAnnotationReducer(current *Document, _ *commands.AnyMessage, payload *document.AddedAnnotationPayload) *Document {
	current.Annotations = document.AddAnnotation(current.Annotations, payload.Annotation)
	return current
}

func removedAnnotationsReducer(current *Document, _ *commands.AnyMessage, payload *document.RemovedAnnotationsPayload) *Document {
	current.Annotations = document.RemoveAnnotations(current.Annotations, payload.AnnotationIDs)
	return current
}

func updatedAnnotationPropsReducer(current *Document, _ *commands.AnyMessage, payload *document.UpdatedAnnotationPropsPayload) *Document {
	current.Annotations = document.UpdateAnnotationProps(current.Annotations, payload.AnnotationIDs, payload.Props)
	return current
}
