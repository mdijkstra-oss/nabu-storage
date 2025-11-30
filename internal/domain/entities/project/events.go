package project

import "hermes-relay/internal/cqrs/commands"

const (
	CreatedProject = "CreatedProject"
	UpdatedProject = "UpdatedProject"
	DeletedProject = "DeletedProject"
	ChangedPhase   = "ChangedPhase"
)

type CreatedProjectPayload = ProjectData
type UpdatedProjectPayload = UpdateProjectData
type DeletedProjectPayload = commands.EmptyPayload
type ChangedPhasePayload = ChangePhasePayload
