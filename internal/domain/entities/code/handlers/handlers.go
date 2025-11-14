package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/project"
	projectview "hermes-relay/internal/domain/projections/project-entity"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/lib/utils"
)

func NewRouter(reg *registry.RegistryState) dispatch.CommandRouter {
	return dispatch.LimitOnEntity(code.EntityName,

		dispatch.LimitOnAction(code.CreateCode,
			registry.Validate(reg, validateCreateCode,
				dispatch.ToCreateEntityEvent[code.CreateCodePayload, code.CreatedCodePayload](code.CreateCode, code.CreatedCode),
			),
		),

		dispatch.LimitOnAction(code.UpdateCode,
			registry.Validate(reg, validateUpdateCode,
				dispatch.ToUpdateEntityEvent[code.UpdateCodePayload, code.UpdatedCodePayload](code.UpdateCode, code.UpdatedCode),
			),
		),

		dispatch.LimitOnAction(code.MergeCodes,
			registry.Validate(reg, validateMergeCodes,
				dispatch.ToUpdateEntityEvent[code.MergeCodesPayload, code.MergedCodesPayload](code.MergeCodes, code.MergedCodes),
			),
		),

		dispatch.ToEmptyDomainEvent(code.DeleteCode, code.DeletedCode),
	)
}

func validateCreateCode(proj project.Project, payload code.CreateCodePayload, msg *commands.AnyMessage) error {
	if !projectview.IsSlugAvailable(proj, payload.Slug, msg.AggregateID) {
		return utils.FieldInUse("slug")
	}

	return nil
}

func validateUpdateCode(proj project.Project, payload code.UpdateCodePayload, msg *commands.AnyMessage) error {
	if payload.Slug != "" {
		if !projectview.IsSlugAvailable(proj, payload.Slug, msg.AggregateID) {
			return utils.FieldInUse("slug")
		}
	}

	return nil
}

func validateMergeCodes(proj project.Project, payload code.MergeCodesPayload, msg *commands.AnyMessage) error {
	if !projectview.CodeExists(proj, payload.SourceID) {
		return utils.FieldNotFound("source_id")
	}

	if !projectview.CodeExists(proj, payload.TargetID) {
		return utils.FieldNotFound("target_id")
	}

	if payload.SourceID == payload.TargetID {
		return utils.FieldError("source_id", "cannot merge with itself")
	}

	return nil
}
