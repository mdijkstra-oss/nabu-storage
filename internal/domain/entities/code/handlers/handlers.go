package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/cqrs/registry"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/lib/utils"
)

func NewRouter(reg *registry.ProjectViewRegistry) dispatch.CommandRouter {
	return dispatch.LimitOnEntity(code.EntityName,

		dispatch.LimitOnAction(code.CreateCode,
			registry.Validate(reg, validateCreateCode,
				dispatch.ToCreateEntityEvent[code.CreateCodePayload](code.CreateCode, code.CreatedCode),
			),
		),

		dispatch.LimitOnAction(code.UpdateCode,
			registry.Validate(reg, validateUpdateCode,
				dispatch.ToUpdateEntityEvent[code.UpdateCodePayload](code.UpdateCode, code.UpdatedCode),
			),
		),

		dispatch.ToEmptyDomainEvent(code.DeleteCode, code.DeletedCode),
	)
}

func validateCreateCode(registry *registry.ProjectView, payload code.CreateCodePayload, msg *commands.AnyMessage) error {
	codes := registry.CodeStore.GetAll()
	if !IsSlugAvailable(codes, payload.Slug, msg.AggregateID) {
		return utils.FieldInUse("slug")
	}

	projects := registry.ProjectStore.GetAll()
	if !projection.EntityExists(projects, payload.ProjectID) {
		return utils.FieldNotFound("project_id")
	}

	return nil
}

func validateUpdateCode(registry *registry.ProjectView, payload code.UpdateCodePayload, msg *commands.AnyMessage) error {
	if payload.Slug != "" {
		codes := registry.CodeStore.GetAll()
		if !IsSlugAvailable(codes, payload.Slug, msg.AggregateID) {
			return utils.FieldInUse("slug")
		}
	}

	return nil
}
