package projection

import "hermes-relay/internal/lib/utils"

type GetByIDQuery struct {
	ID string `path:"id" validate:"required"`
}

type EmptyQuery struct{}

func ByID[T Entity](entities []T, q GetByIDQuery) []T {
	return utils.Filter(entities, func(entity T) bool {
		return entity.GetID() == q.ID
	})
}

func ByAll[T Entity](items []T, _ EmptyQuery) []T {
	return items
}

type PaginationQuery struct {
	Page     int
	PageSize int
}

func Paginate[T Entity](items []T, query PaginationQuery) []T {
	start := (query.Page - 1) * query.PageSize
	if start >= len(items) {
		return []T{}
	}

	end := start + query.PageSize
	if end > len(items) {
		end = len(items)
	}

	return items[start:end]
}
