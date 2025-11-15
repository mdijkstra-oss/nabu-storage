package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
)

func EnsureProjectHealth(registry *RegistryState, handler dispatch.CommandRouter) dispatch.CommandRouter {
	return func(message *commands.AnyMessage, publisher dispatch.PublishFunc) (*commands.AnyMessage, error) {
		projectID := registry.ResolveProjectID(message)
		if projectID == "" {
			return handler(message, publisher)
		}

		proj := registry.GetProject(projectID)
		if proj == nil {
			return nil, utils.FieldError("ProjectID", "not found")
		}

		if !proj.IsHealthy() {
			return nil, &utils.InternalError{Message: "project is in unhealthy state, commands are blocked"}
		}

		return handler(message, publisher)
	}
}

func EnsureEntityHealth(registry *RegistryState, handler dispatch.CommandRouter) dispatch.CommandRouter {
	return func(message *commands.AnyMessage, publisher dispatch.PublishFunc) (*commands.AnyMessage, error) {
		if message.AggregateID == "" {
			return handler(message, publisher)
		}

		projectID := registry.ResolveProjectID(message)
		proj := registry.GetProject(projectID)
		if proj == nil {
			return handler(message, publisher)
		}

		if err := checkEntityHealth(*proj, string(message.AggregateType), message.AggregateID); err != nil {
			return nil, err
		}

		return handler(message, publisher)
	}
}

func checkEntityHealth(proj project.Project, aggregateType, aggregateID string) error {
	switch aggregateType {
	case "Code":
		code, exists := proj.Codes[aggregateID]
		if exists && !code.IsHealthy() {
			return &utils.InternalError{Message: "code is in unhealthy state, commands are blocked"}
		}
	case "File":
		file, exists := proj.Files[aggregateID]
		if exists && !file.IsHealthy() {
			return &utils.InternalError{Message: "file is in unhealthy state, commands are blocked"}
		}
	}
	return nil
}

func ValidateDomain[P any](
	registry *RegistryState,
	validator func(project.Project, P, *commands.AnyMessage) error,
	handler dispatch.CommandRouter,
) dispatch.CommandRouter {
	return func(message *commands.AnyMessage, publisher dispatch.PublishFunc) (*commands.AnyMessage, error) {
		projectID := registry.ResolveProjectID(message)
		proj := registry.GetProject(projectID)

		if proj == nil {
			return nil, utils.FieldError("ProjectID", "not found")
		}

		var payload P
		if err := commands.EnsureValidPayload(message, &payload); err != nil {
			return nil, err
		}

		if err := validator(*proj, payload, message); err != nil {
			return nil, err
		}

		return handler(message, publisher)
	}
}
