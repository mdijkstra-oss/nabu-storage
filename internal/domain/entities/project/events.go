package project

import "hermes-relay/internal/cqrs/commands"

const (
	CreatedProject  = "CreatedProject"
	UpdatedProject  = "UpdatedProject"
	PinnedProject   = "PinnedProject"
	UnpinnedProject = "UnpinnedProject"
	DeletedProject  = "DeletedProject"
	ChangedPhase    = "ChangedPhase"
)

type CreatedProjectPayload = ProjectData
type UpdatedProjectPayload = UpdateProjectData
type PinnedProjectPayload = commands.EmptyPayload
type UnpinnedProjectPayload = commands.EmptyPayload
type DeletedProjectPayload = commands.EmptyPayload
type ChangedPhasePayload = ChangePhasePayload
