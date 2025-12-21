package utils

import "sort"

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

func MapWithIndex[T, U any](slice []T, fn func(int, T) U) []U {
	if slice == nil {
		return nil
	}

	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(i, v)
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

func FindIndex[T any](items []T, predicate func(T) bool) int {
	for i, item := range items {
		if predicate(item) {
			return i
		}
	}
	return -1
}

func Exists[T any](items []T, predicate func(T) bool) bool {
	return Find(items, predicate) != nil
}

func Sort[T any](slice []T, less func(a, b T) bool) {
	sort.Slice(slice, func(i, j int) bool {
		return less(slice[i], slice[j])
	})
}

func Contains[T comparable](slice []T, item T) bool {
	return Exists(slice, func(s T) bool {
		return s == item
	})
}

func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, item := range slice {
		result = fn(result, item)
	}
	return result
}

func ToSet[T comparable](slice []T) map[T]bool {
	set := make(map[T]bool, len(slice))
	for _, item := range slice {
		set[item] = true
	}
	return set
}

func Keys[K comparable, V any](m map[K]V) []K {
	result := make([]K, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
