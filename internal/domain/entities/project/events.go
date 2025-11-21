package project

import "hermes-relay/internal/cqrs/commands"

const (
	CreatedProject = "CreatedProject"
	UpdatedProject = "UpdatedProject"
	DeletedProject = "DeletedProject"
)

type CreatedProjectPayload = ProjectData
type UpdatedProjectPayload = UpdateProjectData
type DeletedProjectPayload = commands.EmptyPayload
