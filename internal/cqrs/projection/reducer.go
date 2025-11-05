package projection

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/utils"
)

type Reducer[T any] func(current *T, event *commands.AnyMessage) *T

func CombineReducers[T any](reducers ...Reducer[T]) Reducer[T] {
	return func(current *T, event *commands.AnyMessage) *T {
		var state = current
		for _, reducer := range reducers {
			state = reducer(state, event)
		}
		return state
	}
}

func For[T any, P any](action commands.Action, reducer func(*T, *commands.AnyMessage, P) *T) Reducer[T] {
	return func(current *T, event *commands.AnyMessage) *T {
		if event.Action != action {
			return current
		}

		var payload P
		utils.MustNotError(commands.UnmarshallPayload(event, &payload))

		return reducer(current, event, payload)
	}
}
