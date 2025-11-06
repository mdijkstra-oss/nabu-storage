package file

import "hermes-relay/internal/cqrs/commands"

const EntityName = "File"

const (
	CreateFile  commands.Action = "CreateFile"
	CodeFile    commands.Action = "CodeFile"
	ClearCoding commands.Action = "ClearCoding" // Remove all coding from given file
)

type CreateFilePayload = CreateFileData

type CodeFilePayload = CodeFileData
