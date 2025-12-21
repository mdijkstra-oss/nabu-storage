package reducer_helpers

import (
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/lib/test-helpers"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
)

type ReducerTestCase[T any] struct {
	Name     string
	Initial  T
	Event    *commands.AnyMessage
	Expected T
}

func RunReducerTests[T any](t *testing.T, tests []ReducerTestCase[T], reducer func(T, *commands.AnyMessage) T, mapFunc ...func(T) T) {
	wrappedReducer := wrapReducerWithImmutabilityCheck(reducer)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := wrappedReducer(tt.Initial, tt.Event)
			actual := result
			if len(mapFunc) > 0 {
				actual = mapFunc[0](result)
			}
			test_helpers.AssertEqual(t, actual, tt.Expected, "result", ignoreMetaFields()...)
		})
	}
}

func ignoreMetaFields() []cmp.Option {
	ignoredFields := map[string]bool{"Version": true, "UpdatedAt": true}
	return []cmp.Option{
		cmp.FilterPath(func(p cmp.Path) bool {
			if sf, ok := p.Last().(cmp.StructField); ok {
				return ignoredFields[sf.Name()]
			}
			return false
		}, cmp.Ignore()),
	}
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

type AggregateChildMapTestConfig[Parent any, Entity any] struct {
	CreatedEvent *commands.AnyMessage
	UpdatedEvent *commands.AnyMessage
	DeletedEvent *commands.AnyMessage

	EntityAfterCreate Entity
	EntityAfterUpdate Entity

	CreateParent func() *Parent
	GetMap       func(*Parent) map[string]Entity
}

func AggregateChildMapTests[Parent any, Entity any](
	config AggregateChildMapTestConfig[Parent, Entity],
) []ReducerTestCase[*Parent] {
	entityID := config.CreatedEvent.AggregateID

	createWithMap := func(entities map[string]Entity) *Parent {
		parent := config.CreateParent()
		childMap := config.GetMap(parent)
		for k, v := range entities {
			childMap[k] = v
		}
		return parent
	}

	return []ReducerTestCase[*Parent]{
		{
			Name:     "Created adds entity to map",
			Initial:  createWithMap(make(map[string]Entity)),
			Event:    config.CreatedEvent,
			Expected: createWithMap(map[string]Entity{entityID: config.EntityAfterCreate}),
		},
		{
			Name:     "Updated modifies entity in map",
			Initial:  createWithMap(map[string]Entity{entityID: config.EntityAfterCreate}),
			Event:    config.UpdatedEvent,
			Expected: createWithMap(map[string]Entity{entityID: config.EntityAfterUpdate}),
		},
		{
			Name:     "Deleted removes entity from map",
			Initial:  createWithMap(map[string]Entity{entityID: config.EntityAfterUpdate}),
			Event:    config.DeletedEvent,
			Expected: createWithMap(make(map[string]Entity)),
		},
	}
}

func wrapReducerWithImmutabilityCheck[T any](reducer func(T, *commands.AnyMessage) T) func(T, *commands.AnyMessage) T {
	if os.Getenv("HERMES_DEV") != "true" {
		return reducer
	}

	return func(current T, event *commands.AnyMessage) T {
		result := reducer(current, event)

		currentVal := reflect.ValueOf(current)
		resultVal := reflect.ValueOf(result)

		if !currentVal.IsValid() || !resultVal.IsValid() {
			return result
		}

		if currentVal.Kind() != reflect.Ptr || resultVal.Kind() != reflect.Ptr {
			return result
		}

		if currentVal.IsNil() || resultVal.IsNil() {
			return result
		}

		if currentVal.Pointer() == resultVal.Pointer() {
			return result
		}

		if projection.HasSharedState(current, result) {
			panic("IMMUTABILITY VIOLATION in test: reducer shares memory between before/after")
		}

		return result
	}
}
