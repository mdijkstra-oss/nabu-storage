package file

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
)

// Router is the combined command router for File entity
var Router = dispatch.CombineRouters(
	dispatch.LimitOnEntity(EntityName,
		dispatch.LimitOnType(commands.Command,
			dispatch.ToCreateEntityEvent[CreatedFilePayload](CreateFile, CreatedFile),
			dispatch.ToUpdateEntityEvent[CodeFilePayload](CodeFile, CodedFile),
			dispatch.ToEmptyDomainEvent(ClearCoding, ClearedCoding),
			dispatch.ToUpdateEntityEvent[MergeCodesPayload](MergeCodes, MergedCodes),
		),
	),
)
