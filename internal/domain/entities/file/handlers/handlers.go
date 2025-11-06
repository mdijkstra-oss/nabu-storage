package handlers

import (
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/file"
)

// Router is the combined command router for File entity
var Router = dispatch.LimitOnEntity(file.EntityName,
	dispatch.ToCreateEntityEvent[file.CreatedFilePayload](file.CreateFile, file.CreatedFile),
	dispatch.ToUpdateEntityEvent[file.CodeFilePayload](file.CodeFile, file.CodedFile),
	dispatch.ToEmptyDomainEvent(file.ClearCoding, file.ClearedCoding),
)
