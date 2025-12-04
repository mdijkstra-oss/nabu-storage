package file

import "hermes-relay/internal/cqrs/commands"

const (
	CreatedFile         = "CreatedFile"
	UpdatedFile         = "UpdatedFile"
	PinnedFile          = "PinnedFile"
	UnpinnedFile        = "UnpinnedFile"
	ReplacedFileContent = "ReplacedFileContent"
	DeletedFile         = "DeletedFile"
	AddedCodeSections   = "AddedCodeSections"
	UpdatedCodeSections = "UpdatedCodeSections"
	RemovedCodeSections = "RemovedCodeSections"
	ClearedCoding       = "ClearedCoding"
	RemovedCodeFromFile = "RemovedCodeFromFile"
)

type CreatedFilePayload struct {
	FileData
	Chunks []Chunk `json:"chunks"`
}

type UpdatedFilePayload = UpdateFilePayload
type PinnedFilePayload = commands.EmptyPayload
type UnpinnedFilePayload = commands.EmptyPayload
type ReplacedFileContentPayload = ReplaceFileContentPayload
type DeletedFilePayload = commands.EmptyPayload
type RemovedCodeFromFilePayload = RemoveCodeFromFilePayload
