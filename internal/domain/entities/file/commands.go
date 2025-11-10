package file

import "hermes-relay/internal/cqrs/commands"

const EntityName = "File"

const (
	CreateFile  commands.Action = "CreateFile"
	UpdateFile  commands.Action = "UpdateFile"
	CodeFile    commands.Action = "CodeFile"
	ClearCoding commands.Action = "ClearCoding" // Remove all coding from given file
)

type CreateFilePayload struct {
	ProjectID   string `json:"project_id" validate:"required"`
	Name        string `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string `json:"description" validate:"max=2000" normalize:"trim"`
	Content     string `json:"content"`
}

type UpdateFilePayload struct {
	Name        string `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string `json:"description" validate:"max=2000" normalize:"trim"`
}

type CodeFilePayload = CodeFileData
