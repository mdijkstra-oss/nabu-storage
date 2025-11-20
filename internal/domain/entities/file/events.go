package file

import "hermes-relay/internal/cqrs/commands"

const (
	CreatedFile              = "CreatedFile"
	UpdatedFile              = "UpdatedFile"
	DeletedFile              = "DeletedFile"
	AddedCodeSections        = "AddedCodeSections"
	UpdatedCodeSections      = "UpdatedCodeSections"
	RemovedCodeSections      = "RemovedCodeSections"
	ClearedCoding            = "ClearedCoding"
)

// CreatedFilePayload Created must always be just entity?
type CreatedFilePayload struct {
	CreateFilePayload
	// Todo will move to create once we can create different types etc
	Type   FileType `json:"type" validate:"omitempty,oneof=codebook source memo context"`
	Locked bool     `json:"locked"`
	Chunks []Chunk  `json:"chunks"`
}

type UpdatedFilePayload = UpdateFilePayload
type DeletedFilePayload = commands.EmptyPayload
