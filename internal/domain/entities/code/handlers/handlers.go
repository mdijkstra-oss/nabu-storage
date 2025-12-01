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
			registry.ValidateDomain(reg, validateCreateCode,
				dispatch.ToCreateEntityEvent[code.CreateCodePayload, code.CreatedCodePayload](code.CreateCode, code.CreatedCode),
			),
		),

		dispatch.LimitOnAction(code.UpdateCode,
			registry.ValidateDomain(reg, validateUpdateCode,
				dispatch.ToUpdateEntityEvent[code.UpdateCodePayload, code.UpdatedCodePayload](code.UpdateCode, code.UpdatedCode),
			),
		),

		dispatch.LimitOnAction(code.MergeCodes,
			registry.ValidateDomain(reg, validateMergeCodes,
				dispatch.ToUpdateEntityEvent[code.MergeCodesPayload, code.MergedCodesPayload](code.MergeCodes, code.MergedCodes),
			),
		),

		dispatch.ToEmptyDomainEvent(code.DeleteCode, code.DeletedCode),
		dispatch.ToEmptyDomainEvent(code.ClearCodeApplications, code.ClearedCodeApplications),

		dispatch.LimitOnAction(code.RecodeAll,
			registry.ValidateDomain(reg, validateRecodeAll,
				dispatch.ToUpdateEntityEvent[code.RecodeAllPayload, code.RecodedAllPayload](code.RecodeAll, code.RecodedAll),
			),
		),
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

func validateRecodeAll(proj project.Project, payload code.RecodeAllPayload, msg *commands.AnyMessage) error {
	if !projectview.CodeExists(proj, payload.TargetCodeID) {
		return utils.FieldNotFound("target_code_id")
	}

	if msg.AggregateID == payload.TargetCodeID {
		return utils.FieldError("target_code_id", "cannot recode to itself")
	}

	return nil
}
