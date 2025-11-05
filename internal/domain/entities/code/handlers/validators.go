package handlers

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/lib/utils"
)

func SlugExists(codes []code.Code, slug, excludeID string) bool {
	_, err := utils.Find(codes, func(c code.Code) bool {
		return c.Slug == slug && c.ID != excludeID
	})
	return err == nil
}

func ValidateUniqueSlug(codes []code.Code, slug, excludeID string) error {
	if SlugExists(codes, slug, excludeID) {
		return utils.MakeValidationFieldError("Slug", "slug already exists")
	}
	return nil
}
