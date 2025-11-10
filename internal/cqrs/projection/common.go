package projection

import (
	"hermes-relay/internal/cqrs/commands"
	"reflect"
)

type ProjectChild interface {
	GetProjectID() string
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
