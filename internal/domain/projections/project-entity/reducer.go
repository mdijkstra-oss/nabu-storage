package projectview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	codeview "hermes-relay/internal/domain/projections/code-entity"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	"hermes-relay/internal/lib/utils"
)

var Reducer = projection.WithHealthCheck(
	projection.CombineReducers(
		projection.For(project.CreatedProject, CreatedProjectReducer),
		projection.IfExists(
			projection.For(project.UpdatedProject, UpdatedProjectReducer),
			projection.For(project.DeletedProject, projection.DeletedEntity[Project]),
			projection.ApplyChildReducerToMap(
				func(p *Project) map[string]code.Code { return p.Codes },
				codeview.Reducer,
			),
			projection.ApplyChildReducerToMap(
				func(p *Project) map[string]file.File { return p.Files },
				fileview.Reducer,
			),
		),
	),
)

func CreatedProjectReducer(_ *Project, message *commands.AnyMessage, payload *project.CreatedProjectPayload) *Project {
	return &Project{
		ID:          message.AggregateID,
		Name:        payload.Name,
		Description: payload.Description,
		Codes:       make(map[string]code.Code),
		Files:       make(map[string]file.File),
		Healthy:     true,
	}
}

func UpdatedProjectReducer(current *Project, _ *commands.AnyMessage, payload *project.UpdatedProjectPayload) *Project {
	updated := utils.ApplyPartialUpdate(*current, payload)
	return &updated
}
