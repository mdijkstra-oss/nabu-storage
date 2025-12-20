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
