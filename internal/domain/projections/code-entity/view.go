package codeview

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/lib/utils"
	"hermes-relay/internal/projection"
)

type Code = code.Code

type SlugQuery struct {
	Slug string `path:"slug" validate:"required,code_slug"`
}

func GetBySlug(store *projection.Store[Code], q SlugQuery) (*Code, error) {
	found, err := utils.Find(store.GetAll(), func(c Code) bool {
		return c.Slug == q.Slug
	})

	if err != nil {
		return nil, err
	}

	return &found, nil
}
