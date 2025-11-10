package project

import "hermes-relay/internal/cqrs/commands"

const EntityName = "Project"

const (
	CreateProject = "CreateProject"
	UpdateProject = "UpdateProject"
	DeleteProject = "DeleteProject"
)

type CreateProjectPayload = CreateProjectData
type UpdateProjectPayload = UpdateProjectData
type DeleteProjectPayload = commands.EmptyPayload
