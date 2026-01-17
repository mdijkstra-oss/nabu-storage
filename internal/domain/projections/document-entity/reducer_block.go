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

func toBlockTree(d *Document) document.BlockTree {
	if d.Blocks == nil {
		return document.NewBlockTree()
	}
	return document.BlockTree{
		Blocks: d.Blocks,
		HeadID: d.HeadID,
		TailID: d.TailID,
	}
}

func applyBlockTree(d *Document, tree document.BlockTree) *Document {
	d.Blocks = tree.Blocks
	d.HeadID = tree.HeadID
	d.TailID = tree.TailID
	return d
}

func insertedBlocksReducer(current *Document, _ *commands.AnyMessage, payload *document.InsertedBlocksPayload) *Document {
	tree := toBlockTree(current)
	tree, _ = document.InsertBlocksAfter(tree, payload.Position, payload.Blocks)
	return applyBlockTree(current, tree)
}

func deletedBlocksReducer(current *Document, _ *commands.AnyMessage, payload *document.DeletedBlocksPayload) *Document {
	tree := toBlockTree(current)
	tree = document.DeleteBlocksByID(tree, payload.BlockIDs)
	return applyBlockTree(current, tree)
}

func replacedBlocksReducer(current *Document, _ *commands.AnyMessage, payload *document.ReplacedBlocksPayload) *Document {
	tree := toBlockTree(current)
	tree = document.ReplaceBlocksByID(tree, payload.BlockIDs, payload.Blocks)
	return applyBlockTree(current, tree)
}

func movedBlocksReducer(current *Document, _ *commands.AnyMessage, payload *document.MovedBlocksPayload) *Document {
	tree := toBlockTree(current)
	tree = document.MoveBlocksAfter(tree, payload.BlockIDs, payload.Position)
	return applyBlockTree(current, tree)
}

func replacedContentReducer(current *Document, _ *commands.AnyMessage, payload *document.ReplacedContentPayload) *Document {
	tree := document.FromArray(payload.Content)
	return applyBlockTree(current, tree)
}

func updatedBlockPropsReducer(current *Document, _ *commands.AnyMessage, payload *document.UpdatedBlockPropsPayload) *Document {
	tree := toBlockTree(current)
	tree = document.UpdateBlocksProps(tree, payload.BlockIDs, payload.Props)
	return applyBlockTree(current, tree)
}
