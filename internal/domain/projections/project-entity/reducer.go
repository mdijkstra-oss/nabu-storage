package projectview

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/document"
	"hermes-relay/internal/domain/entities/project"
	documentview "hermes-relay/internal/domain/projections/document-entity"
)

var projectReducer = projection.WithTimestamp[Project](
	projection.WithVersionIncrement(
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
	),
)

var Reducer = projection.WithImmutabilityCheck(
	projection.CombineReducers(
		projectReducer,
		projection.IfExists(
			projection.ApplyChildReducerToMap(
				func(p *Project) map[string]document.Document { return p.Documents },
				func(p *Project, documents map[string]document.Document) *Project {
					updated := *p
					updated.Documents = documents
					return &updated
				},
				documentview.Reducer,
			),
		),
	),
)

func CreatedProjectReducer(_ *Project, message *commands.AnyMessage, payload *project.CreatedProjectPayload) *Project {
	return &Project{
		ID:          message.AggregateID,
		Healthy:     true,
		ProjectData: *payload,
		Documents:   make(map[string]document.Document),
	}
}
