package handlers

import (
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/registry"
	"hermes-relay/internal/domain/entities/file"
)

func NewRouter(_ *registry.ProjectViewRegistry) dispatch.CommandRouter {
	return dispatch.LimitOnEntity(file.EntityName,
		dispatch.ToCreateEntityEvent[file.CreateFilePayload, file.CreatedFilePayload](file.CreateFile, file.CreatedFile, func(payload *file.CreateFilePayload) file.CreatedFilePayload {
			return file.CreatedFilePayload{
				CreateFilePayload: file.CreateFilePayload{
					ProjectID:   payload.ProjectID,
					Name:        payload.Name,
					Description: payload.Description,
					Content:     payload.Content,
				},
				Type:   file.FileTypeSource,
				Locked: true,
			}
		}),
		dispatch.ToUpdateEntityEvent[file.UpdateFilePayload, file.UpdatedFilePayload](file.UpdateFile, file.UpdatedFile),
		dispatch.ToUpdateEntityEvent[file.CodeFilePayload, file.CodedFilePayload](file.CodeFile, file.CodedFile),
		dispatch.ToEmptyDomainEvent(file.ClearCoding, file.ClearedCoding),
	)
}
