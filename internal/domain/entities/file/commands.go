package file

import "hermes-relay/internal/cqrs/commands"

const EntityName = "File"

const (
	CreateFile  commands.Action = "CreateFile"
	CodeFile    commands.Action = "CodeFile"
	ClearCoding commands.Action = "ClearCoding" // Remove all coding from given file
)

type CreateFilePayload struct {
	ProjectID string `json:"project_id" validate:"required"`
	Name      string `json:"name" validate:"required,max=255" normalize:"trim"`
	Content   string `json:"content"`
}

type CodeFilePayload = CodeFileData
