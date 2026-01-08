package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
)

func EnsureHealth(store *Store, handler dispatch.CommandRouter) dispatch.CommandRouter {
	return func(message *commands.AnyMessage, publisher dispatch.PublishFunc) (*commands.AnyMessage, error) {
		if message.AggregateID == "" {
			return handler(message, publisher)
		}

		projectID := projection.Read(store, func(r *Registry) string {
			return ResolveProjectID(r, message)
		})
		if projectID == "" {
			return handler(message, publisher)
		}

		proj := projection.Read(store, func(r *Registry) *project.Project {
			return GetProject(r, projectID)
		})
		if proj == nil {
			if isCreateAction(message.Action) {
				return handler(message, publisher)
			}
			return nil, utils.FieldError("ProjectID", "not found")
		}

		if !proj.IsHealthy() {
			return nil, &utils.InternalError{Message: "project is in unhealthy state, commands are blocked"}
		}

		if err := checkEntityHealth(*proj, message.Action, string(message.AggregateType), message.AggregateID); err != nil {
			return nil, err
		}

		return handler(message, publisher)
	}
}

func EnsureExpectedVersion(store *Store, handler dispatch.CommandRouter) dispatch.CommandRouter {
	return func(message *commands.AnyMessage, publisher dispatch.PublishFunc) (*commands.AnyMessage, error) {
		if message.ExpectedEntityVersion == nil {
			return handler(message, publisher)
		}

		proj := projection.Read(store, func(r *Registry) *project.Project {
			projectID := ResolveProjectID(r, message)
			return GetProject(r, projectID)
		})
		if proj == nil {
			return handler(message, publisher)
		}

		if err := checkExpectedVersion(*proj, message); err != nil {
			return nil, err
		}

		return handler(message, publisher)
	}
}

func checkExpectedVersion(proj project.Project, message *commands.AnyMessage) error {
	expected := *message.ExpectedEntityVersion

	var entity projection.Versionable
	var exists bool
	switch message.AggregateType {
	case "Project":
		entity, exists = proj, true
	case "Document":
		entity, exists = proj.GetDocument(message.AggregateID)
	default:
		panic("checkExpectedVersion: unknown aggregate type: " + string(message.AggregateType))
	}

	if !exists {
		return nil
	}
	if expected != entity.GetVersion() {
		return utils.VersionConflict(expected, entity.GetVersion())
	}
	return nil
}

func isCreateAction(action commands.Action) bool {
	return len(action) >= 6 && action[:6] == "Create"
}

func checkEntityHealth(proj project.Project, action commands.Action, aggregateType, aggregateID string) error {
	var entity projection.Healthable
	var exists bool
	switch aggregateType {
	case "Project":
		entity, exists = proj, true
	case "Document":
		entity, exists = proj.GetDocument(aggregateID)
	default:
		panic("checkEntityHealth: unknown aggregate type: " + aggregateType)
	}

	if !exists {
		if isCreateAction(action) {
			return nil
		}
		return utils.FieldError("aggregate_id", aggregateType+" not found")
	}
	if !entity.IsHealthy() {
		return &utils.InternalError{Message: aggregateType + " is in unhealthy state, commands are blocked"}
	}
	return nil
}

func ValidateDomain[P any](
	store *Store,
	validator func(project.Project, P, *commands.AnyMessage) error,
	handler dispatch.CommandRouter,
) dispatch.CommandRouter {
	return NormalizeDomain[P](
		store,
		func(proj project.Project, payload P, msg *commands.AnyMessage) (P, error) {
			return payload, validator(proj, payload, msg)
		},
		handler,
	)
}

func NormalizeDomain[P any](
	store *Store,
	normalize func(project.Project, P, *commands.AnyMessage) (P, error),
	handler dispatch.CommandRouter,
) dispatch.CommandRouter {
	return TransformDomain(store, normalize, handler)
}

func TransformDomain[In, Out any](
	store *Store,
	transform func(project.Project, In, *commands.AnyMessage) (Out, error),
	handler dispatch.CommandRouter,
) dispatch.CommandRouter {
	return func(message *commands.AnyMessage, publisher dispatch.PublishFunc) (*commands.AnyMessage, error) {
		proj := projection.Read(store, func(r *Registry) *project.Project {
			projectID := ResolveProjectID(r, message)
			return GetProject(r, projectID)
		})

		if proj == nil {
			return nil, utils.FieldError("ProjectID", "not found")
		}

		var payload In
		if err := commands.EnsureValidPayload(message, &payload); err != nil {
			return nil, err
		}

		transformedPayload, err := transform(*proj, payload, message)
		if err != nil {
			return nil, err
		}

		if err := utils.ToValidationError(utils.Validate.Struct(transformedPayload)); err != nil {
			return nil, err
		}

		updatedMessage := *message
		updatedMessage.Payload = transformedPayload

		return handler(&updatedMessage, publisher)
	}
}
