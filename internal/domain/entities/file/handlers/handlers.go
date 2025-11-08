package handlers

import (
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/registry"
	"hermes-relay/internal/domain/entities/file"
)

func NewRouter(_ *registry.ProjectViewRegistry) dispatch.CommandRouter {
	return dispatch.LimitOnEntity(file.EntityName,
		dispatch.ToCreateEntityEvent[file.CreatedFilePayload](file.CreateFile, file.CreatedFile, func(payload *file.CreateFileData) {
			// Set defaults for now, not really relevant but nice to already store for later
			payload.Type = file.FileTypeSource
			payload.Locked = true

			// Original defaults to empty string (zero value)
		}),
		dispatch.ToUpdateEntityEvent[file.CodeFilePayload](file.CodeFile, file.CodedFile),
		dispatch.ToEmptyDomainEvent(file.ClearCoding, file.ClearedCoding),
	)
}
