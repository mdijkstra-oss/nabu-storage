package codeview

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/lib/utils"
)

type Code = code.Code

func GetBySlug(codes []code.Code, slug string) *code.Code {
	results := utils.Filter(codes, func(c code.Code) bool {
		return c.Slug == slug
	})
	if len(results) == 0 {
		return nil
	}
	return &results[0]
}

func IsSlugAvailable(codes []code.Code, slug string, excludeID string) bool {
	return !utils.Exists(codes, func(c code.Code) bool {
		return c.Slug == slug && c.ID != excludeID
	})
}
