package code

import (
	"hermes-relay/internal/cqrs/commands"
)

const EntityName commands.AggregateType = "Code"

const (
	CreateCode commands.Action = "CreateCode"
	UpdateCode commands.Action = "UpdateCode"
	DeleteCode commands.Action = "DeleteCode"
)

type CreateCodePayload = CreateCodeData
type UpdateCodePayload = UpdateCodeData
