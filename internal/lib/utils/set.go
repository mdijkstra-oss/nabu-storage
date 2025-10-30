package utils

import "errors"

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

func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(slice)) // Pre-allocate with capacity

	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}

	return result
}

func Find[T any](items []T, predicate func(T) bool) (T, error) {
	for _, item := range items {
		if predicate(item) {
			return item, nil
		}
	}
	var zero T
	return zero, errors.New("item not found")
}
