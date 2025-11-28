package file

import "hermes-relay/internal/cqrs/commands"

const (
	CreatedFile         = "CreatedFile"
	UpdatedFile         = "UpdatedFile"
	UpdatedFileContent  = "UpdatedFileContent"
	DeletedFile         = "DeletedFile"
	AddedCodeSections   = "AddedCodeSections"
	UpdatedCodeSections = "UpdatedCodeSections"
	RemovedCodeSections = "RemovedCodeSections"
	ClearedCoding       = "ClearedCoding"
)

type CreatedFilePayload struct {
	FileData
	Chunks []Chunk `json:"chunks"`
}

type UpdatedFilePayload = UpdateFilePayload
type UpdatedFileContentPayload = UpdateFileContentPayload
type DeletedFilePayload = commands.EmptyPayload
