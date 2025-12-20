package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/lib/utils"
)

func NewBlockRouter() dispatch.CommandRouter {
	return dispatch.LimitOnEntity(document.EntityName,
		withBlockValidation[document.InsertBlocksPayload](document.InsertBlocks, document.InsertedBlocks, getBlocksFromInsert),
		dispatch.ToUpdateEntityEvent[document.DeleteBlocksPayload, document.DeletedBlocksPayload](document.DeleteBlocks, document.DeletedBlocks),
		withBlockValidation[document.ReplaceBlocksPayload](document.ReplaceBlocks, document.ReplacedBlocks, getBlocksFromReplace),
		dispatch.ToUpdateEntityEvent[document.MoveBlocksPayload, document.MovedBlocksPayload](document.MoveBlocks, document.MovedBlocks),
		withBlockValidation[document.ReplaceContentPayload](document.ReplaceContent, document.ReplacedContent, getBlocksFromContent),
	)
}

func getBlocksFromInsert(p *document.InsertBlocksPayload) []document.Block {
	return p.Blocks
}

func getBlocksFromReplace(p *document.ReplaceBlocksPayload) []document.Block {
	return p.Blocks
}

func getBlocksFromContent(p *document.ReplaceContentPayload) []document.Block {
	return p.Content
}

func withBlockValidation[P any](commandAction, eventAction commands.Action, getBlocks func(*P) []document.Block) dispatch.CommandRouter {
	return func(message *commands.AnyMessage, publisher dispatch.PublishFunc) (*commands.AnyMessage, error) {
		if message.Action != commandAction {
			return nil, nil
		}

		var payload P
		if err := commands.EnsureValidPayload(message, &payload); err != nil {
			return nil, err
		}

		blocks := getBlocks(&payload)
		if err := document.ValidateBlocks(blocks); err != nil {
			return nil, utils.FieldError("blocks", err.Error())
		}

		return commands.ToDomainEvent(message, eventAction), nil
	}
}
