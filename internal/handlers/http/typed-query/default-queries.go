package typedquery

import (
	"hermes-relay/internal/projection"
)

type GetByIDQuery struct {
	ID string `path:"id" validate:"required"`
}

type EmptyQuery struct{}

func GetById[T any](store *projection.Store[T], q GetByIDQuery) (*T, error) {
	file, err := store.GetByID(q.ID)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func GetAll[T any](store *projection.Store[T], q EmptyQuery) ([]T, error) {
	all := store.GetAll()
	return all, nil
}
