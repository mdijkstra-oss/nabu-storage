package utils

func Values[K comparable, V any](m map[K]V) []V {
	result := make([]V, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

func Map[T, U any](slice []T, fn func(T) U) []U {
	if slice == nil {
		return nil
	}

	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func FlatMap[T any, U any](slice []T, fn func(T) []U) []U {
	result := []U{}
	for _, item := range slice {
		result = append(result, fn(item)...)
	}
	return result
}

func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(slice)) // Pre-allocate with capacity

	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}

	return result
}

func Find[T any](items []T, predicate func(T) bool) *T {
	for _, item := range items {
		if predicate(item) {
			return &item
		}
	}
	return nil
}

func Exists[T any](items []T, predicate func(T) bool) bool {
	return Find(items, predicate) != nil
}

func Sort[T any](slice []T, less func(a, b T) bool) {
	for i := 0; i < len(slice); i++ {
		for j := i + 1; j < len(slice); j++ {
			if !less(slice[i], slice[j]) {
				slice[i], slice[j] = slice[j], slice[i]
			}
		}
	}
}
