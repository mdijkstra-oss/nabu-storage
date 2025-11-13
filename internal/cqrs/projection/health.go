package projection

import "hermes-relay/internal/cqrs/commands"

type Healthable interface {
	MarkUnhealthy()
	IsHealthy() bool
}

type healthablePtr[T any] interface {
	*T
	Healthable
}

func WithHealthCheck[T any, PT healthablePtr[T]](reducer Reducer[T]) Reducer[T] {
	return func(current *T, event *commands.AnyMessage) (result *T) {
		defer func() {
			if r := recover(); r != nil {
				if current != nil {
					PT(current).MarkUnhealthy()
					result = current
				}
			}
		}()
		return reducer(current, event)
	}
}
