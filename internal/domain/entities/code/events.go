package code

import (
	"hermes-relay/internal/cqrs/commands"
)

const (
	CreatedCode commands.Action = "CreatedCode"
	UpdatedCode commands.Action = "UpdatedCode"
	DeletedCode commands.Action = "DeletedCode"
)

type CreatedCodePayload = CreateCodeData
type UpdatedCodePayload = UpdateCodeData
type DeletedCodePayload = DeleteCodeData
