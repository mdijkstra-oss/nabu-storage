package project

import "hermes-relay/internal/cqrs/commands"

const EntityName commands.AggregateType = "Project"

const (
	CreateProject commands.Action = "CreateProject"
	UpdateProject commands.Action = "UpdateProject"
	PinProject    commands.Action = "PinProject"
	UnpinProject  commands.Action = "UnpinProject"
	DeleteProject commands.Action = "DeleteProject"
	ChangePhase   commands.Action = "ChangePhase"
)

type CreateProjectPayload = CreateProjectData
type UpdateProjectPayload = UpdateProjectData
type PinProjectPayload = commands.EmptyPayload
type UnpinProjectPayload = commands.EmptyPayload
type DeleteProjectPayload = commands.EmptyPayload

type ChangePhasePayload struct {
	Phase Phase `json:"phase" validate:"required,oneof=explore code validate analyze"`
}
