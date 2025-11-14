package projection

import "hermes-relay/internal/cqrs/commands"

type Healthable interface {
	IsHealthy() bool
	WithUnhealthy() any
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
					marked := PT(current).WithUnhealthy()
					result = marked.(*T)
				}
			}
		}()
		return reducer(current, event)
	}
}
