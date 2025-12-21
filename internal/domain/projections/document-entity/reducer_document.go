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
			Content:     []document.Block{},
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
