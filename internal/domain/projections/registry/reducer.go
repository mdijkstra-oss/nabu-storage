package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	projectview "hermes-relay/internal/domain/projections/project-entity"
)

var Reducer = NewReducer(projectview.Reducer)

func NewReducer(projectReducer projection.Reducer[project.Project]) projection.Reducer[Registry] {
	return func(current *Registry, event *commands.AnyMessage) *Registry {
		if current == nil {
			current = &Registry{
				Projects:        make(map[string]project.Project),
				EntityToProject: make(map[string]string),
			}
		}

		projectID := extractProjectID(current, event)
		if projectID == "" {
			return current
		}

		proj, exists := current.Projects[projectID]
		var projPtr *project.Project
		if exists {
			projPtr = &proj
		} else {
			projPtr = nil
		}

		newProj := projectReducer(projPtr, event)

		if newProj == nil {
			delete(current.Projects, projectID)
		} else {
			current.Projects[projectID] = *newProj
		}

		updateLookupTable(current.EntityToProject, event, projectID)

		return current
	}
}

func extractProjectID(registry *Registry, event *commands.AnyMessage) string {
	if event.AggregateType == project.EntityName {
		return event.AggregateID
	}

	projectID := commands.ExtractProjectID(event)
	if projectID != "" {
		return projectID
	}

	key := string(event.AggregateType) + ":" + event.AggregateID
	return registry.EntityToProject[key]
}

func updateLookupTable(lookup map[string]string, event *commands.AnyMessage, projectID string) {
	if event.AggregateType != code.EntityName && event.AggregateType != file.EntityName {
		return
	}

	key := string(event.AggregateType) + ":" + event.AggregateID

	if commands.IsCreatedEvent(event.Action) {
		lookup[key] = projectID
	} else if commands.IsDeletedEvent(event.Action) {
		delete(lookup, key)
	}
}
