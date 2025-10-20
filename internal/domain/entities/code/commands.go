package code

import "hermes-relay/internal/cqrs"

const (
	CreateCode cqrs.Action = "CreateCode"
	UpdateCode cqrs.Action = "UpdateCode"
	DeleteCode cqrs.Action = "DeleteCode"
)

type CreateCodePayload = CreateCodeData
type UpdateCodePayload = UpdateCodeData
