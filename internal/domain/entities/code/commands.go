package code

import (
	"hermes-relay/internal/cqrs/commands"
)

const EntityName commands.AggregateType = "Code"

const (
	CreateCode             commands.Action = "CreateCode"
	UpdateCode             commands.Action = "UpdateCode"
	DeleteCode             commands.Action = "DeleteCode"
	MergeCodes             commands.Action = "MergeCodes"
	ClearCodeApplications  commands.Action = "ClearCodeApplications"
	RecodeAll              commands.Action = "RecodeAll"
)

type CreateCodePayload = CodeData
type UpdateCodePayload = UpdateCodeData
type DeleteCodePayload = commands.EmptyPayload
type MergeCodesPayload = MergeCodesData
type ClearCodeApplicationsPayload = commands.EmptyPayload
type RecodeAllPayload = RecodeAllData
