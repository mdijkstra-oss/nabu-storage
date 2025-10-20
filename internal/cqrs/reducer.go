package cqrs

import "hermes-relay/internal/utils"

type Reducer[T any] func(current *T, event *Message) *T

func CombineReducers[T any](reducers ...Reducer[T]) Reducer[T] {
	return func(current *T, event *Message) *T {
		var state = current
		for _, reducer := range reducers {
			state = reducer(state, event)
		}
		return state
	}
}

func For[T any, P any](action Action, reducer func(*T, *Message, P) *T) Reducer[T] {
	return func(current *T, event *Message) *T {
		if event.Action != action {
			return current
		}

		var payload P
		utils.MustNotError(UnmarshallPayload(event, &payload))

		return reducer(current, event, payload)
	}
}
