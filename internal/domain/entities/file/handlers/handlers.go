package handlers

import (
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/registry"
	"hermes-relay/internal/domain/entities/file"
)

func NewRouter(_ *registry.ProjectViewRegistry) dispatch.CommandRouter {
	return dispatch.LimitOnEntity(file.EntityName,
		dispatch.ToCreateEntityEvent[file.CreatedFilePayload](file.CreateFile, file.CreatedFile),
		dispatch.ToUpdateEntityEvent[file.CodeFilePayload](file.CodeFile, file.CodedFile),
		dispatch.ToEmptyDomainEvent(file.ClearCoding, file.ClearedCoding),
	)
}
