package projectview

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
)

var Reducer = cqrs.CombineReducers(
	cqrs.For(project.CreatedProject, CreatedProjectReducer),
	cqrs.For(file.CreatedFile, FileCreatedReducer),
	cqrs.For(code.CreatedCode, CodeCreatedReducer),
	cqrs.For(code.DeletedCode, CodeDeletedReducer),
)

func CreatedProjectReducer(_ *Project, message *cqrs.AnyMessage, payload *project.CreatedProjectPayload) *Project {
	return &Project{
		ID:      message.AggregateID,
		Name:    payload.Name,
		CodeIDs: []string{},
		FileIDs: []string{},
	}
}

func FileCreatedReducer(current *Project, message *cqrs.AnyMessage, payload *file.CreatedFilePayload) *Project {
	if current == nil || current.ID != payload.ProjectID {
		return current
	}

	current.FileIDs = append(current.FileIDs, message.AggregateID)
	return current
}

func CodeCreatedReducer(current *Project, message *cqrs.AnyMessage, payload *code.CreatedCodePayload) *Project {
	if current == nil || current.ID != payload.ProjectID {
		return current
	}

	current.CodeIDs = append(current.CodeIDs, message.AggregateID)
	return current
}

func CodeDeletedReducer(current *Project, message *cqrs.AnyMessage, payload any) *Project {
	if current == nil {
		return current
	}

	current.CodeIDs = utils.Filter(current.CodeIDs, func(id string) bool {
		return id != message.AggregateID
	})
	return current
}
