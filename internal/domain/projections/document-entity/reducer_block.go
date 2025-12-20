package documentview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/document"
)

var BlockReducer = projection.IfExists(
	projection.For(document.InsertedBlocks, insertedBlocksReducer),
	projection.For(document.DeletedBlocks, deletedBlocksReducer),
	projection.For(document.ReplacedBlocks, replacedBlocksReducer),
	projection.For(document.MovedBlocks, movedBlocksReducer),
	projection.For(document.ReplacedContent, replacedContentReducer),
	projection.For(document.UpdatedBlockProps, updatedBlockPropsReducer),
)

func insertedBlocksReducer(current *Document, _ *commands.AnyMessage, payload *document.InsertedBlocksPayload) *Document {
	content, _ := document.InsertBlocksAfter(current.Content, payload.Position, payload.Blocks)
	current.Content = content
	return current
}

func deletedBlocksReducer(current *Document, _ *commands.AnyMessage, payload *document.DeletedBlocksPayload) *Document {
	current.Content = document.DeleteBlocksByID(current.Content, payload.BlockIDs)
	return current
}

func replacedBlocksReducer(current *Document, _ *commands.AnyMessage, payload *document.ReplacedBlocksPayload) *Document {
	current.Content = document.ReplaceBlocksByID(current.Content, payload.BlockIDs, payload.Blocks)
	return current
}

func movedBlocksReducer(current *Document, _ *commands.AnyMessage, payload *document.MovedBlocksPayload) *Document {
	current.Content = document.MoveBlocksAfter(current.Content, payload.BlockIDs, payload.Position)
	return current
}

func replacedContentReducer(current *Document, _ *commands.AnyMessage, payload *document.ReplacedContentPayload) *Document {
	current.Content = payload.Content
	return current
}

func updatedBlockPropsReducer(current *Document, _ *commands.AnyMessage, payload *document.UpdatedBlockPropsPayload) *Document {
	current.Content = document.UpdateBlocksProps(current.Content, payload.BlockIDs, payload.Props)
	return current
}
