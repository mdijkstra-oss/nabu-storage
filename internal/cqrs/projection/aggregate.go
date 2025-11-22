package projection

import "hermes-relay/internal/cqrs/commands"

func ApplyChildReducerToMap[Parent Entity, Child Entity](
	getMap func(*Parent) map[string]Child,
	setMap func(*Parent, map[string]Child) *Parent,
	childReducer Reducer[Child],
) Reducer[Parent] {
	return func(current *Parent, event *commands.AnyMessage) *Parent {
		if current == nil {
			return nil
		}

		entityMap := getMap(current)
		entityID := event.AggregateID

		entity, exists := entityMap[entityID]
		var entityPtr *Child
		if exists {
			entityPtr = &entity
		} else {
			entityPtr = nil
		}

		newEntity := childReducer(entityPtr, event)

		// Create new map to avoid shared state between before/after snapshots.
		// Bootstrap captures state before and after event application for patch generation.
		// If we mutate the original map, both snapshots point to same data = null patches.
		newMap := copyMap(entityMap)

		if newEntity == nil {
			delete(newMap, entityID)
		} else if (*newEntity).GetID() != "" {
			newMap[entityID] = *newEntity
		}

		return setMap(current, newMap)
	}
}

func copyMap[K comparable, V any](m map[K]V) map[K]V {
	newMap := make(map[K]V, len(m))
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}
