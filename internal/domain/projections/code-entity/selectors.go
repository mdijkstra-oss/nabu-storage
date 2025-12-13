package codeview

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/lib/utils"
)

type Code = code.Code

func IsSlugAvailable(codes []code.Code, slug string, excludeID string) bool {
	return !utils.Exists(codes, func(c code.Code) bool {
		return c.Slug == slug && c.ID != excludeID
	})
}

func FindCodeBySlug(codes []code.Code, slug string) *code.Code {
	for i := range codes {
		if codes[i].Slug == slug {
			return &codes[i]
		}
	}
	return nil
}
