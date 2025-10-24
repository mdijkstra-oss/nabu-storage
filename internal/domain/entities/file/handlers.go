package file

import (
	"hermes-relay/internal/cqrs"
)

// Router is the combined command router for File entity
var Router = cqrs.CombineRouters(
	cqrs.LimitOnEntity(EntityName,
		cqrs.LimitOnType(cqrs.Command,
			cqrs.ToCreateEntityEvent[CreatedFilePayload](Create, CreatedFile),
			cqrs.ToUpdateEntityEvent[CodeFilePayload](CodeFile, CodedFile),
			cqrs.ToEmptyDomainEvent(ClearCoding, ClearedCoding),
		),
	),
)
