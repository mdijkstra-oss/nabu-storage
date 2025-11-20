package code

import (
	"hermes-relay/internal/cqrs/commands"
)

const (
	CreatedCode commands.Action = "CreatedCode"
	UpdatedCode commands.Action = "UpdatedCode"
	DeletedCode commands.Action = "DeletedCode"
	MergedCodes commands.Action = "MergedCodes"
)

type CreatedCodePayload = CreateCodeData
type UpdatedCodePayload = UpdateCodeData
type DeletedCodePayload = commands.EmptyPayload
type MergedCodesPayload = MergeCodesData
