package code

import (
	"hermes-relay/internal/cqrs/commands"
)

const (
	CreatedCode             commands.Action = "CreatedCode"
	UpdatedCode             commands.Action = "UpdatedCode"
	DeletedCode             commands.Action = "DeletedCode"
	MergedCodes             commands.Action = "MergedCodes"
	ClearedCodeApplications commands.Action = "ClearedCodeApplications"
	RecodedAll              commands.Action = "RecodedAll"
)

type CreatedCodePayload = CodeData
type UpdatedCodePayload = UpdateCodeData
type DeletedCodePayload = commands.EmptyPayload
type MergedCodesPayload = MergeCodesData
type ClearedCodeApplicationsPayload = commands.EmptyPayload
type RecodedAllPayload = RecodeAllData
