package projection

import "hermes-relay/internal/cqrs/commands"

type Versionable interface {
	GetVersion() int
	WithVersion(int) any
}

type versionablePtr[T any] interface {
	*T
	Versionable
}

func WithVersionIncrement[T any, PT versionablePtr[T]](reducer Reducer[T]) Reducer[T] {
	return func(current *T, event *commands.AnyMessage) *T {
		result := reducer(current, event)
		if result == nil || result == current {
			return result
		}
		currentVersion := 0
		if current != nil {
			currentVersion = PT(current).GetVersion()
		}
		versioned := PT(result).WithVersion(currentVersion + 1)
		return versioned.(*T)
	}
}
