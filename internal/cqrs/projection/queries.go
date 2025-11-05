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
	Page     int `query:"page" validate:"min=1"`
	PageSize int `query:"page_size" validate:"min=1,max=100"`
}

type PaginationResult[T any] struct {
	Items    []T `json:"items"`
	Total    int `json:"total"`
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

func Paginate[T Entity](items []T, query PaginationQuery) []PaginationResult[T] {
	total := len(items)
	start := (query.Page - 1) * query.PageSize

	// Handle out of bounds
	if start >= total {
		return []PaginationResult[T]{{
			Items:    []T{},
			Total:    total,
			Page:     query.Page,
			PageSize: query.PageSize,
		}}
	}

	end := start + query.PageSize
	if end > total {
		end = total
	}

	return []PaginationResult[T]{{
		Items:    items[start:end],
		Total:    total,
		Page:     query.Page,
		PageSize: query.PageSize,
	}}
}
