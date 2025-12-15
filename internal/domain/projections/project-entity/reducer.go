package projectview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/chart"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	chartview "hermes-relay/internal/domain/projections/chart-entity"
	codeview "hermes-relay/internal/domain/projections/code-entity"
	fileview "hermes-relay/internal/domain/projections/file-entity"
)

var projectReducer = projection.WithVersionIncrement(
	projection.WithHealthCheck(
		projection.CombineReducers(
			projection.For(project.CreatedProject, CreatedProjectReducer),
			projection.IfExists(
				projection.For(project.UpdatedProject, projection.UpdatedEntity[Project, project.UpdatedProjectPayload]),
				projection.For(project.PinnedProject, projection.PinnedEntity[Project]),
				projection.For(project.UnpinnedProject, projection.UnpinnedEntity[Project]),
				projection.For(project.DeletedProject, projection.DeletedEntity[Project]),
			),
		),
	),
)

var Reducer = projection.WithImmutabilityCheck(
	projection.CombineReducers(
		projectReducer,
		projection.IfExists(
			projection.ApplyChildReducerToMap(
				func(p *Project) map[string]chart.Chart { return p.Charts },
				func(p *Project, charts map[string]chart.Chart) *Project {
					updated := *p
					updated.Charts = charts
					return &updated
				},
				chartview.Reducer,
			),
			projection.ApplyChildReducerToMap(
				func(p *Project) map[string]code.Code { return p.Codes },
				func(p *Project, codes map[string]code.Code) *Project {
					updated := *p
					updated.Codes = codes
					return &updated
				},
				codeview.Reducer,
			),
			projection.ApplyChildReducerToMap(
				func(p *Project) map[string]file.File { return p.Files },
				func(p *Project, files map[string]file.File) *Project {
					updated := *p
					updated.Files = files
					return &updated
				},
				fileview.Reducer,
			),
		),
	),
)

func CreatedProjectReducer(_ *Project, message *commands.AnyMessage, payload *project.CreatedProjectPayload) *Project {
	return &Project{
		ID:          message.AggregateID,
		Healthy:     true,
		ProjectData: *payload,
		Charts:      make(map[string]chart.Chart),
		Codes:       make(map[string]code.Code),
		Files:       make(map[string]file.File),
	}
}
