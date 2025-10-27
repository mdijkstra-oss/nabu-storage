package code

import "hermes-relay/internal/cqrs"

const (
	CreatedCode cqrs.Action = "CreatedCode"
	UpdatedCode cqrs.Action = "UpdatedCode"
	DeletedCode cqrs.Action = "DeletedCode"
)

type CreatedCodePayload = CreateCodeData
type UpdatedCodePayload = UpdateCodeData
type DeletedCodePayload = DeleteCodeData
