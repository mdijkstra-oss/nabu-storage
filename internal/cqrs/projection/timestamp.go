package projection

import (
	"hermes-relay/internal/cqrs/commands"
	"time"
)

type Timestampable interface {
	WithUpdatedAt(time.Time) any
}

type timestampablePtr[T any] interface {
	*T
	Timestampable
}

func WithTimestamp[T any, PT timestampablePtr[T]](reducer Reducer[T]) Reducer[T] {
	return func(current *T, event *commands.AnyMessage) *T {
		result := reducer(current, event)
		if result == nil || result == current {
			return result
		}
		return PT(result).WithUpdatedAt(event.Timestamp).(*T)
	}
}
