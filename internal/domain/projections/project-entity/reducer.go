package projectview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
)

var Reducer = projection.CombineReducers(
	projection.For(project.CreatedProject, CreatedProjectReducer),
	projection.IfExists(
		projection.For(project.UpdatedProject, UpdatedProjectReducer),
		projection.For(project.DeletedProject, projection.DeletedEntity[Project]),
		// Todo: Not called properly somehow, ordering?
		projection.For(file.CreatedFile, AddedFileToProjectReducer),
		projection.For(file.DeletedFile, RemovedCodeFileProjectReducer),
		projection.For(code.CreatedCode, AddedCodeToProjectReducer),
		projection.For(code.DeletedCode, RemovedCodeFromProjectReducer),
	),
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

func UpdatedProjectReducer(current *Project, message *commands.AnyMessage, payload *project.UpdatedProjectPayload) *Project {
	current.Name = payload.Name
	if payload.Description != "" {
		current.Description = payload.Description
	}
	return current
}

// External Events

// Todo: Not called for some reason, but that is fine
func AddedFileToProjectReducer(current *Project, message *commands.AnyMessage, _ any) *Project {
	current.FileIDs = append(current.FileIDs, message.AggregateID)
	return current
}

func AddedCodeToProjectReducer(current *Project, message *commands.AnyMessage, _ any) *Project {
	current.CodeIDs = append(current.CodeIDs, message.AggregateID)
	return current
}

func RemovedCodeFromProjectReducer(current *Project, message *commands.AnyMessage, _ any) *Project {
	current.CodeIDs = utils.Filter(current.CodeIDs, func(id string) bool {
		return id != message.AggregateID
	})
	return current
}

func RemovedCodeFileProjectReducer(current *Project, message *commands.AnyMessage, _ any) *Project {
	current.CodeIDs = utils.Filter(current.CodeIDs, func(id string) bool {
		return id != message.AggregateID
	})
	return current
}
