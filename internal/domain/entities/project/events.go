package project

const (
	CreatedProject         = "CreatedProject"
	AddedFileToProject     = "AddedFileToProject"
	AddedCodeToProject     = "AddedCodeToProject"
	RemovedCodeFromProject = "RemovedCodeFromProject"
)

type CreatedProjectPayload = CreateProjectData
type AddedFileToProjectPayload = AddedFileToProjectData
type AddedCodeToProjectPayload = AddedCodeToProjectData
type RemovedCodeFromProjectPayload = RemovedCodeFromProjectData
