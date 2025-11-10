package projection

import (
	"hermes-relay/internal/cqrs/commands"
	"reflect"
)

type ProjectChild interface {
	GetProjectID() string
}

// DeletedEntity is a generic reducer for entity deletion events.
// It returns nil to remove the entity from the projection.
func DeletedEntity[T any](_ *T, _ *commands.AnyMessage, _ any) *T {
	return nil
}

func DeletedProjectReducer[T ProjectChild](current *T, message *commands.AnyMessage) *T {
	if message.Action != "DeletedProject" {
		return current
	}

	if current == nil {
		return nil
	}

	v := reflect.ValueOf(current)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}

	if (*current).GetProjectID() == message.AggregateID {
		return nil
	}

	return current
}
