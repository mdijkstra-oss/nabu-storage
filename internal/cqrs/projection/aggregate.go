package projection

import "hermes-relay/internal/cqrs/commands"

func ApplyChildReducerToMap[Parent Entity, Child Entity](
	getMap func(*Parent) map[string]Child,
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

		if newEntity == nil {
			delete(entityMap, entityID)
		} else if (*newEntity).GetID() != "" {
			entityMap[entityID] = *newEntity
		}

		return current
	}
}
