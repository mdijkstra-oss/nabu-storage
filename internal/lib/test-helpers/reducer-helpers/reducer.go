package reducer_helpers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/utils"
	"testing"
)

type ReducerTestCase[T any] struct {
	Name     string
	Initial  T
	Event    *commands.AnyMessage
	Expected T
}

type reducerInput[T any] struct {
	initial T
	event   *commands.AnyMessage
}

func RunReducerTests[T any](t *testing.T, tests []ReducerTestCase[T], reducer func(T, *commands.AnyMessage) T, mapFunc ...func(T) T) {
	genericTests := utils.Map(tests, func(tt ReducerTestCase[T]) struct {
		Name     string
		Input    reducerInput[T]
		Expected T
	} {
		return struct {
			Name     string
			Input    reducerInput[T]
			Expected T
		}{
			Name: tt.Name,
			Input: reducerInput[T]{
				initial: tt.Initial,
				event:   tt.Event,
			},
			Expected: tt.Expected,
		}
	})

	test_helpers.RunFunctionTests(t, genericTests, func(input reducerInput[T]) T {
		return reducer(input.initial, input.event)
	}, mapFunc...)
}
