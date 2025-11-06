package handlers

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/lib/utils"
)

func IsSlugAvailable(codes []code.Code, slug, excludeID string) bool {
	return !utils.Exists(codes, func(c code.Code) bool {
		return c.Slug == slug && c.ID != excludeID
	})
}
