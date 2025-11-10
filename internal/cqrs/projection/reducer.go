package projection

import (
	"fmt"
	"hermes-relay/internal/cqrs/commands"
	"log/slog"
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

func IfExists[T any](reducers ...Reducer[T]) Reducer[T] {
	combined := CombineReducers(reducers...)
	return func(current *T, event *commands.AnyMessage) *T {
		if current == nil {
			return nil
		}
		return combined(current, event)
	}
}

func For[T any, P any](action commands.Action, reducer func(*T, *commands.AnyMessage, P) *T) Reducer[T] {
	return func(current *T, event *commands.AnyMessage) *T {
		if event.Action != action {
			return current
		}

		var payload P
		if err := commands.UnmarshallPayload(event, &payload); err != nil {
			slog.Error("FATAL: failed to unmarshal event payload",
				"action", action,
				"aggregateType", event.AggregateType,
				"aggregateID", event.AggregateID,
				"timestamp", event.Timestamp,
				"error", err)
			panic(fmt.Sprintf("corrupt %s event for %s (ChunkID: %s): %v",
				action, event.AggregateType, event.AggregateID, err))
		}

		return reducer(current, event, payload)
	}
}
