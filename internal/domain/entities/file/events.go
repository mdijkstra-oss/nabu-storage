package file

import "hermes-relay/internal/cqrs/commands"

const (
	CreatedFile          = "CreatedFile"
	UpdatedFile          = "UpdatedFile"
	PinnedFile           = "PinnedFile"
	UnpinnedFile         = "UnpinnedFile"
	ReplacedFileContent  = "ReplacedFileContent"
	DeletedFile          = "DeletedFile"
	ClearedCoding        = "ClearedCoding"
	RemovedCodeFromFile  = "RemovedCodeFromFile"
	ModifiedCodeSections = "ModifiedCodeSections"
)

type CreatedFilePayload struct {
	FileData
	Content string         `json:"content"`
	Codes   []CodedSection `json:"codes"`
}

type UpdatedFilePayload = UpdateFilePayload
type PinnedFilePayload = commands.EmptyPayload
type UnpinnedFilePayload = commands.EmptyPayload
type ReplacedFileContentPayload = ReplaceFileContentPayload
type DeletedFilePayload = commands.EmptyPayload
type RemovedCodeFromFilePayload = RemoveCodeFromFilePayload
