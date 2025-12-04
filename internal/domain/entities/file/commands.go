package file

import "hermes-relay/internal/cqrs/commands"

const EntityName commands.AggregateType = "File"

const (
	CreateFile         commands.Action = "CreateFile"
	UpdateFile         commands.Action = "UpdateFile"
	PinFile            commands.Action = "PinFile"
	UnpinFile          commands.Action = "UnpinFile"
	ReplaceFileContent commands.Action = "ReplaceFileContent"
	EditFileContent    commands.Action = "EditFileContent"
	DeleteFile         commands.Action = "DeleteFile"
	AddCodeSections    commands.Action = "AddCodeSections"
	UpdateCodeSections commands.Action = "UpdateCodeSections"
	RemoveCodeSections commands.Action = "RemoveCodeSections"
	ClearCoding        commands.Action = "ClearCoding"
	RemoveCodeFromFile commands.Action = "RemoveCodeFromFile"
)

type CreateFilePayload struct {
	ProjectID   string   `json:"project_id" validate:"required,valid_id"`
	Name        string   `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string   `json:"description" validate:"max=2000" normalize:"trim"`
	Content     string   `json:"content"`
	Type        FileType `json:"type" validate:"omitempty,oneof=corpus codebook memo llm-memo"`
}

type UpdateFilePayload struct {
	Name        string `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string `json:"description,omitempty" validate:"omitempty,max=2000" normalize:"trim"`
}

type ReplaceFileContentPayload struct {
	Content string `json:"content" validate:"required"`
}

type EditFileContentPayload struct {
	OldText string `json:"old_text" validate:"required"`
	NewText string `json:"new_text" validate:"required"`
}

type PinFilePayload = commands.EmptyPayload
type UnpinFilePayload = commands.EmptyPayload
type DeleteFilePayload = commands.EmptyPayload

type RemoveCodeFromFilePayload struct {
	CodeID string `json:"code_id" validate:"required,valid_id"`
}
