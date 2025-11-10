package reducer_helpers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
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

// DeletedEntityTests generates table-driven tests for entity deletion.
// It tests that an entity is removed from the projection when deleted.
func DeletedEntityTests[T any](
	entityName commands.AggregateType,
	deletedAction commands.Action,
	createEntity func() *T,
) []ReducerTestCase[*T] {
	return []ReducerTestCase[*T]{
		{
			Name:     string(deletedAction) + " deletes existing entity",
			Initial:  createEntity(),
			Event:    domain_helpers.NewDomainEvent(entityName, "entity-1", deletedAction, nil),
			Expected: nil,
		},
		{
			Name:     string(deletedAction) + " on nil state returns nil",
			Initial:  nil,
			Event:    domain_helpers.NewDomainEvent(entityName, "entity-1", deletedAction, nil),
			Expected: nil,
		},
	}
}

func DeletedProjectCascadeTests[T projection.ProjectChild](
	createEntity func(projectID string) *T,
) []ReducerTestCase[*T] {
	return []ReducerTestCase[*T]{
		{
			Name:     "DeletedProject cascades to entity with matching ProjectID",
			Initial:  createEntity("project-1"),
			Event:    domain_helpers.NewDomainEvent("Project", "project-1", "DeletedProject", nil),
			Expected: nil,
		},
		{
			Name:     "DeletedProject does not affect entity with different ProjectID",
			Initial:  createEntity("project-2"),
			Event:    domain_helpers.NewDomainEvent("Project", "project-1", "DeletedProject", nil),
			Expected: createEntity("project-2"),
		},
		{
			Name:     "DeletedProject on nil entity returns nil",
			Initial:  nil,
			Event:    domain_helpers.NewDomainEvent("Project", "project-1", "DeletedProject", nil),
			Expected: nil,
		},
	}
}
