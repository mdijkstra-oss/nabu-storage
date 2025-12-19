package document

import "hermes-relay/internal/cqrs/commands"

const (
	CreatedDocument  = "CreatedDocument"
	UpdatedDocument  = "UpdatedDocument"
	PinnedDocument   = "PinnedDocument"
	UnpinnedDocument = "UnpinnedDocument"
	DeletedDocument  = "DeletedDocument"

	InsertedBlocks  = "InsertedBlocks"
	DeletedBlocks   = "DeletedBlocks"
	ReplacedBlocks  = "ReplacedBlocks"
	MovedBlocks     = "MovedBlocks"
	ReplacedContent = "ReplacedContent"
)

type CreatedDocumentPayload struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdatedDocumentPayload = UpdateDocumentPayload
type PinnedDocumentPayload = commands.EmptyPayload
type UnpinnedDocumentPayload = commands.EmptyPayload
type DeletedDocumentPayload = commands.EmptyPayload

type InsertedBlocksPayload = InsertBlocksPayload
type DeletedBlocksPayload = DeleteBlocksPayload
type ReplacedBlocksPayload = ReplaceBlocksPayload
type MovedBlocksPayload = MoveBlocksPayload
type ReplacedContentPayload = ReplaceContentPayload
