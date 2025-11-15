package projection

type EmptyQuery struct{}

type IDQuery struct {
	ID string `path:"id" validate:"required,valid_id"`
}

type SlugQuery struct {
	Slug string `path:"slug" validate:"required,code_slug"`
}

type PaginationQuery struct {
	Page     int `query:"page" validate:"min=1" default:"1"`
	PageSize int `query:"page_size" validate:"min=1,max=100" default:"20"`
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
