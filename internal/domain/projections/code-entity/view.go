package codeview

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/lib/utils"
)

type Code = code.Code

type SlugQuery struct {
	Slug string `path:"slug" validate:"required,code_slug"`
}

func BySlug(code []Code, query SlugQuery) []Code {
	return utils.Filter(code, func(item Code) bool {
		return item.Slug == query.Slug
	})
}
