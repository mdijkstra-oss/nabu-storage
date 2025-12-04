package projection

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/utils"
	"reflect"
)

type ProjectChild interface {
	GetProjectID() string
}

type Pinnable interface {
	WithPinned(bool) any
}

// DeletedEntity is a generic reducer for entity deletion events.
// It returns nil to remove the entity from the projection.
func DeletedEntity[T any](_ *T, _ *commands.AnyMessage, _ any) *T {
	return nil
}

func PinnedEntity[T Pinnable](current *T, _ *commands.AnyMessage, _ *commands.EmptyPayload) *T {
	return (*current).WithPinned(true).(*T)
}

func UnpinnedEntity[T Pinnable](current *T, _ *commands.AnyMessage, _ *commands.EmptyPayload) *T {
	return (*current).WithPinned(false).(*T)
}

func UpdatedEntity[T, P any](current *T, _ *commands.AnyMessage, payload *P) *T {
	updated := utils.ApplyPartialUpdate(*current, payload)
	return &updated
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
