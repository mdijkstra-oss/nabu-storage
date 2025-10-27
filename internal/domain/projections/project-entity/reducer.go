package projectview

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
)

var Reducer = cqrs.CombineReducers(
	cqrs.For(project.CreatedProject, CreatedProjectReducer),
	cqrs.For(project.AddedFileToProject, AddedFileToProjectReducer),
	cqrs.For(project.AddedCodeToProject, AddedCodeToProjectReducer),
	cqrs.For(project.RemovedCodeFromProject, RemovedCodeFromProjectReducer),
)

func CreatedProjectReducer(_ *Project, message *cqrs.AnyMessage, payload *project.CreatedProjectPayload) *Project {
	return &Project{
		ID:      message.AggregateID,
		Name:    payload.Name,
		CodeIDs: []string{},
		FileIDs: []string{},
	}
}

func AddedFileToProjectReducer(current *Project, message *cqrs.AnyMessage, payload *project.AddedFileToProjectPayload) *Project {
	if current == nil {
		return current
	}

	current.FileIDs = append(current.FileIDs, payload.FileID)
	return current
}

func AddedCodeToProjectReducer(current *Project, message *cqrs.AnyMessage, payload *project.AddedCodeToProjectPayload) *Project {
	if current == nil {
		return current
	}

	current.CodeIDs = append(current.CodeIDs, payload.CodeID)
	return current
}

func RemovedCodeFromProjectReducer(current *Project, message *cqrs.AnyMessage, payload *project.RemovedCodeFromProjectPayload) *Project {
	if current == nil {
		return current
	}

	current.CodeIDs = utils.Filter(current.CodeIDs, func(id string) bool {
		return id != payload.CodeID
	})
	return current
}
