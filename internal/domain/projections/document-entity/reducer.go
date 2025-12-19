package documentview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/lib/utils"
	"time"
)

var Reducer = projection.WithVersionIncrement(
	projection.WithHealthCheck(
		projection.CombineReducers(
			projection.For(document.CreatedDocument, CreatedDocumentReducer),
			projection.IfExists(
				projection.For(document.UpdatedDocument, projection.UpdatedEntity[Document, document.UpdatedDocumentPayload]),
				projection.For(document.PinnedDocument, projection.PinnedEntity[Document]),
				projection.For(document.UnpinnedDocument, projection.UnpinnedEntity[Document]),
				projection.For(document.DeletedDocument, projection.DeletedEntity[Document]),
				projection.For(document.InsertedBlocks, InsertedBlocksReducer),
				projection.For(document.DeletedBlocks, DeletedBlocksReducer),
				projection.For(document.ReplacedBlocks, ReplacedBlocksReducer),
				projection.For(document.MovedBlocks, MovedBlocksReducer),
				projection.For(document.ReplacedContent, ReplacedContentReducer),
			),
			projection.DeletedProjectReducer[document.Document],
		),
	),
)

func CreatedDocumentReducer(_ *Document, message *commands.AnyMessage, payload *document.CreatedDocumentPayload) *Document {
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

func findInsertIndex(blocks []document.Block, position string) int {
	if position == "" {
		return 0
	}
	for i, b := range blocks {
		if b.ID == position {
			return i + 1
		}
	}
	return len(blocks)
}

func insertBlocksAt(blocks []document.Block, index int, newBlocks []document.Block) []document.Block {
	result := make([]document.Block, 0, len(blocks)+len(newBlocks))
	result = append(result, blocks[:index]...)
	result = append(result, newBlocks...)
	result = append(result, blocks[index:]...)
	return result
}

func isInSet(id string, ids []string) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}

func InsertedBlocksReducer(current *Document, _ *commands.AnyMessage, payload *document.InsertedBlocksPayload) *Document {
	index := findInsertIndex(current.Content, payload.Position)
	current.Content = insertBlocksAt(current.Content, index, payload.Blocks)
	return current
}

func DeletedBlocksReducer(current *Document, _ *commands.AnyMessage, payload *document.DeletedBlocksPayload) *Document {
	current.Content = utils.Filter(current.Content, func(b document.Block) bool {
		return !isInSet(b.ID, payload.BlockIDs)
	})
	return current
}

func ReplacedBlocksReducer(current *Document, _ *commands.AnyMessage, payload *document.ReplacedBlocksPayload) *Document {
	result := make([]document.Block, 0, len(current.Content))
	replaced := false
	for _, b := range current.Content {
		if isInSet(b.ID, payload.BlockIDs) {
			if !replaced {
				result = append(result, payload.Blocks...)
				replaced = true
			}
		} else {
			result = append(result, b)
		}
	}
	current.Content = result
	return current
}

func MovedBlocksReducer(current *Document, _ *commands.AnyMessage, payload *document.MovedBlocksPayload) *Document {
	moving := utils.Filter(current.Content, func(b document.Block) bool {
		return isInSet(b.ID, payload.BlockIDs)
	})
	remaining := utils.Filter(current.Content, func(b document.Block) bool {
		return !isInSet(b.ID, payload.BlockIDs)
	})
	index := findInsertIndex(remaining, payload.Position)
	current.Content = insertBlocksAt(remaining, index, moving)
	return current
}

func ReplacedContentReducer(current *Document, _ *commands.AnyMessage, payload *document.ReplacedContentPayload) *Document {
	current.Content = payload.Content
	return current
}
