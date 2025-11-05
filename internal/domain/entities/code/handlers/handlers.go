package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/registry"
	"hermes-relay/internal/domain/entities/code"
)

func NewRouter(reg *registry.ProjectViewRegistry) dispatch.CommandRouter {
	return dispatch.CombineRouters(
		dispatch.LimitOnEntity(code.EntityName,
			dispatch.LimitOnType(commands.Command,
				registry.Validate(reg, validateCreateCode,
					dispatch.ToCreateEntityEvent[code.CreateCodePayload](code.CreateCode, code.CreatedCode),
				),
				registry.Validate(reg, validateUpdateCode,
					dispatch.ToUpdateEntityEvent[code.UpdateCodePayload](code.UpdateCode, code.UpdatedCode),
				),

				dispatch.ToEmptyDomainEvent(code.DeleteCode, code.DeletedCode),
			),
		),
	)
}

func validateCreateCode(registry *registry.ProjectView, payload code.CreateCodePayload, _ *commands.AnyMessage) error {
	codes := registry.CodeStore.GetAll()
	return ValidateUniqueSlug(codes, payload.Slug, "")
}

func validateUpdateCode(registry *registry.ProjectView, payload code.CreateCodePayload, msg *commands.AnyMessage) error {
	codes := registry.CodeStore.GetAll()
	return ValidateUniqueSlug(codes, payload.Slug, msg.AggregateID)
}
