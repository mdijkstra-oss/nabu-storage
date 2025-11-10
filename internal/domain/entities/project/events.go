package project

import "hermes-relay/internal/cqrs/commands"

const (
	CreatedProject         = "CreatedProject"
	UpdatedProject         = "UpdatedProject"
	DeletedProject         = "DeletedProject"
	AddedFileToProject     = "AddedFileToProject"
	AddedCodeToProject     = "AddedCodeToProject"
	RemovedCodeFromProject = "RemovedCodeFromProject"
)

type CreatedProjectPayload = CreateProjectData
type UpdatedProjectPayload = UpdateProjectData
type DeletedProjectPayload = commands.EmptyPayload

type AddedFileToProjectPayload = AddedFileToProjectData
type AddedCodeToProjectPayload = AddedCodeToProjectData
type RemovedCodeFromProjectPayload = RemovedCodeFromProjectData
