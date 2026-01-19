package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/lib/utils"
)

func NewBlockRouter(store *registry.Store) dispatch.CommandRouter {
	return dispatch.LimitOnEntity(document.EntityName,
		withBlockValidation[document.InsertBlocksPayload](document.InsertBlocks, document.InsertedBlocks, assignAndGetBlocksFromInsert),
		dispatch.ToUpdateEntityEvent[document.DeleteBlocksPayload, document.DeletedBlocksPayload](document.DeleteBlocks, document.DeletedBlocks),
		dispatch.ToUpdateEntityEvent[document.MoveBlocksPayload, document.MovedBlocksPayload](document.MoveBlocks, document.MovedBlocks),
		withBlockValidation[document.ReplaceContentPayload](document.ReplaceContent, document.ReplacedContent, assignAndGetBlocksFromContent),
		registry.ValidateDomain(store, validateUpdateBlock,
			dispatch.ToUpdateEntityEvent[document.UpdateBlockPayload, document.UpdatedBlockPayload](document.UpdateBlock, document.UpdatedBlock),
		),
	)
}

func validateUpdateBlock(proj project.Project, payload document.UpdateBlockPayload, msg *commands.AnyMessage) error {
	doc, exists := proj.GetDocument(msg.AggregateID)
	if !exists {
		return utils.FieldError("aggregate_id", "document not found")
	}

	tree := document.ToBlockTree(doc.DocumentData)
	if _, found := document.FindBlock(tree, payload.BlockID); !found {
		return utils.FieldError("block_id", "block not found: "+payload.BlockID)
	}

	return nil
}

func assignAndGetBlocksFromInsert(p *document.InsertBlocksPayload) []document.Block {
	p.Blocks = document.AssignBlockIDs(p.Blocks)
	return p.Blocks
}

func assignAndGetBlocksFromContent(p *document.ReplaceContentPayload) []document.Block {
	p.Content = document.AssignBlockIDs(p.Content)
	return p.Content
}

func withBlockValidation[P any](commandAction, eventAction commands.Action, assignAndGetBlocks func(*P) []document.Block) dispatch.CommandRouter {
	return func(message *commands.AnyMessage, publisher dispatch.PublishFunc) (*commands.AnyMessage, error) {
		if message.Action != commandAction {
			return nil, nil
		}

		var payload P
		if err := commands.EnsureValidPayload(message, &payload); err != nil {
			return nil, err
		}

		blocks := assignAndGetBlocks(&payload)
		if err := document.ValidateBlocks(blocks); err != nil {
			return nil, utils.FieldError("blocks", err.Error())
		}

		return commands.ToDomainEvent(message, eventAction, any(payload)), nil
	}
}
