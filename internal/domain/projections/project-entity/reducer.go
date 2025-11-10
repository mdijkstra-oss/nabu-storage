package projectview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
)

var Reducer = projection.CombineReducers(
	projection.For(project.CreatedProject, CreatedProjectReducer),
	projection.For(project.UpdatedProject, UpdatedProjectReducer),
	projection.For(project.DeletedProject, DeletedProjectReducer),
	projection.For(project.AddedFileToProject, AddedFileToProjectReducer),
	projection.For(project.AddedCodeToProject, AddedCodeToProjectReducer),
	projection.For(project.RemovedCodeFromProject, RemovedCodeFromProjectReducer),
)

func CreatedProjectReducer(_ *Project, message *commands.AnyMessage, payload *project.CreatedProjectPayload) *Project {
	return &Project{
		ID:          message.AggregateID,
		Name:        payload.Name,
		Description: payload.Description,
		CodeIDs:     []string{},
		FileIDs:     []string{},
	}
}

func AddedFileToProjectReducer(current *Project, message *commands.AnyMessage, payload *project.AddedFileToProjectPayload) *Project {
	if current == nil {
		return current
	}

	current.FileIDs = append(current.FileIDs, payload.FileID)
	return current
}

func AddedCodeToProjectReducer(current *Project, message *commands.AnyMessage, payload *project.AddedCodeToProjectPayload) *Project {
	if current == nil {
		return current
	}

	current.CodeIDs = append(current.CodeIDs, payload.CodeID)
	return current
}

func UpdatedProjectReducer(current *Project, message *commands.AnyMessage, payload *project.UpdatedProjectPayload) *Project {
	if current == nil {
		return current
	}

	current.Name = payload.Name
	current.Description = payload.Description
	return current
}

func DeletedProjectReducer(_ *Project, _ *commands.AnyMessage, _ any) *Project {
	return nil
}

func RemovedCodeFromProjectReducer(current *Project, message *commands.AnyMessage, payload *project.RemovedCodeFromProjectPayload) *Project {
	if current == nil {
		return current
	}

	current.CodeIDs = utils.Filter(current.CodeIDs, func(id string) bool {
		return id != payload.CodeID
	})
	return current
}
