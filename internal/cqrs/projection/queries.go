package projection

type EmptyQuery struct{}

type IDQuery struct {
	ID string `path:"id" validate:"required,valid_id"`
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

type CursorQuery struct {
	SinceID   string `query:"since_id"`
	Limit     int    `query:"limit" validate:"min=1,max=100" default:"20"`
	ActorType string `query:"actor_type" validate:"omitempty,oneof=human llm system"`
}

type CursorResult[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor"`
	HasMore    bool   `json:"has_more"`
}

type Identifiable interface {
	GetID() string
}

type ActorTyped interface {
	GetActorType() string
}

func CursorFilter[T Identifiable](items []T, query CursorQuery, getActorType func(T) string) CursorResult[T] {
	filtered := filterBySinceID(items, query.SinceID)
	filtered = filterByActorType(filtered, query.ActorType, getActorType)

	hasMore := len(filtered) > query.Limit
	if hasMore {
		filtered = filtered[:query.Limit]
	}

	nextCursor := resolveNextCursor(items, filtered, query.SinceID)

	return CursorResult[T]{
		Items:      filtered,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}

func filterBySinceID[T Identifiable](items []T, sinceID string) []T {
	if sinceID == "" {
		return items
	}
	for i, item := range items {
		if item.GetID() == sinceID {
			return items[i+1:]
		}
	}
	return items
}

func filterByActorType[T any](items []T, actorType string, getActorType func(T) string) []T {
	if actorType == "" {
		return items
	}
	var result []T
	for _, item := range items {
		if getActorType(item) == actorType {
			result = append(result, item)
		}
	}
	return result
}

func resolveNextCursor[T Identifiable](allItems []T, filtered []T, sinceID string) string {
	if len(filtered) > 0 {
		return filtered[len(filtered)-1].GetID()
	}
	if len(allItems) > 0 {
		return allItems[len(allItems)-1].GetID()
	}
	return sinceID
}
