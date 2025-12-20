package document

import "hermes-relay/internal/cqrs/commands"

const EntityName commands.AggregateType = "Document"

const (
	PositionHead = "head"
	PositionTail = "tail"
)

const (
	CreateDocument commands.Action = "CreateDocument"
	UpdateDocument commands.Action = "UpdateDocument"
	PinDocument    commands.Action = "PinDocument"
	UnpinDocument  commands.Action = "UnpinDocument"
	DeleteDocument commands.Action = "DeleteDocument"

	InsertBlocks     commands.Action = "InsertBlocks"
	DeleteBlocks     commands.Action = "DeleteBlocks"
	ReplaceBlocks    commands.Action = "ReplaceBlocks"
	MoveBlocks       commands.Action = "MoveBlocks"
	ReplaceContent   commands.Action = "ReplaceContent"
	UpdateBlockProps commands.Action = "UpdateBlockProps"
)

type CreateDocumentPayload struct {
	ProjectID   string `json:"project_id" validate:"required,project_id" normalize:"project_id"`
	Name        string `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string `json:"description" validate:"max=2000" normalize:"trim"`
}

type UpdateDocumentPayload struct {
	Name        string `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string `json:"description,omitempty" validate:"omitempty,max=2000" normalize:"trim"`
}

type PinDocumentPayload = commands.EmptyPayload
type UnpinDocumentPayload = commands.EmptyPayload
type DeleteDocumentPayload = commands.EmptyPayload

type InsertBlocksPayload struct {
	Position string  `json:"position" validate:"required,block_position"`
	Blocks   []Block `json:"blocks" validate:"required,min=1"`
}

type DeleteBlocksPayload struct {
	BlockIDs []string `json:"block_ids" validate:"required,min=1"`
}

type ReplaceBlocksPayload struct {
	BlockIDs []string `json:"block_ids" validate:"required,min=1"`
	Blocks   []Block  `json:"blocks" validate:"required,min=1"`
}

type MoveBlocksPayload struct {
	BlockIDs []string `json:"block_ids" validate:"required,min=1"`
	Position string   `json:"position" validate:"required,block_position"`
}

type ReplaceContentPayload struct {
	Content []Block `json:"content" validate:"required"`
}

type UpdateBlockPropsPayload struct {
	BlockIDs []string   `json:"block_ids" validate:"required,min=1"`
	Props    BlockProps `json:"props"`
}
