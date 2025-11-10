package file

import "hermes-relay/internal/cqrs/commands"

const (
	CreatedFile   = "CreatedFile"
	UpdatedFile   = "UpdatedFile"
	DeletedFile   = "DeletedFile"
	CodedFile     = "CodedFile"
	ClearedCoding = "ClearedCoding"
)

// CreatedFilePayload Created must always be just entity?
type CreatedFilePayload struct {
	CreateFilePayload
	// Todo will move to create once we can create different types etc
	Type FileType `json:"type" validate:"omitempty,oneof=codebook source memo context"`
	// Locked is done @ command to event time
	Locked bool `json:"locked"` // whether file is read-only
}

type UpdatedFilePayload = UpdateFilePayload
type DeletedFilePayload = commands.EmptyPayload

type CodedFilePayload = CodeFileData
