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
	Type   FileType `json:"type" validate:"omitempty,oneof=codebook source memo context"`
	Locked bool     `json:"locked"` // whether file is read-only
	Chunks []Chunk  `json:"chunks"` // pre-chunked content
}

type UpdatedFilePayload = UpdateFilePayload
type DeletedFilePayload = commands.EmptyPayload

type CodedFilePayload = CodeFileData
