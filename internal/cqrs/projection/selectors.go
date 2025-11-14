package projection

import "hermes-relay/internal/lib/utils"

func GetByID[T Entity](entities []T, id string) *T {
	for _, e := range entities {
		if e.GetID() == id {
			return &e
		}
	}
	return nil
}

func EntityExists[T Entity](entities []T, id string) bool {
	return utils.Exists(entities, func(e T) bool {
		return e.GetID() == id
	})
}

func GetFromMap[K comparable, V any](m map[K]V, key K) *V {
	v, exists := m[key]
	if !exists {
		return nil
	}
	return &v
}

func ExistsInMap[K comparable, V any](m map[K]V, key K) bool {
	_, exists := m[key]
	return exists
}
