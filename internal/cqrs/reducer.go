package cqrs

import (
	"hermes-relay/internal/lib/utils"
)

type Reducer[T any] func(current *T, event *AnyMessage) *T

func CombineReducers[T any](reducers ...Reducer[T]) Reducer[T] {
	return func(current *T, event *AnyMessage) *T {
		var state = current
		for _, reducer := range reducers {
			state = reducer(state, event)
		}
		return state
	}
}

func For[T any, P any](action Action, reducer func(*T, *AnyMessage, P) *T) Reducer[T] {
	return func(current *T, event *AnyMessage) *T {
		if event.Action != action {
			return current
		}

		var payload P
		utils.MustNotError(UnmarshallPayload(event, &payload))

		return reducer(current, event, payload)
	}
}
