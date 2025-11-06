package test_helpers

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"hermes-relay/internal/cqrs/commands"
	"reflect"
	"testing"
)

func AssertMessage(t *testing.T, got, want *commands.AnyMessage, msg string) {
	t.Helper()
	AssertEqualIgnoreFields(t, got, want, msg, commands.AnyMessage{}, "ID", "Timestamp", "CausationID")
}

func AssertEqual(t *testing.T, got, want any, msg string, opts ...cmp.Option) {
	t.Helper()

	// Handle nil comparisons using reflection to catch typed nils
	gotVal := reflect.ValueOf(got)
	wantVal := reflect.ValueOf(want)
	gotIsNil := !gotVal.IsValid() || (gotVal.Kind() == reflect.Ptr && gotVal.IsNil())
	wantIsNil := !wantVal.IsValid() || (wantVal.Kind() == reflect.Ptr && wantVal.IsNil())

	if gotIsNil && wantIsNil {
		return
	}

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Fatalf("%s: mismatch (-want +got):\n%s", msg, diff)
	}
}

func AssertEqualIgnoreFields(t *testing.T, got, want any, msg string, ignoreType any, fields ...string) {
	t.Helper()
	AssertEqual(t, got, want, msg, cmpopts.IgnoreFields(ignoreType, fields...))
}

func AssertError(t *testing.T, got error, wantMsg string, msg string) {
	t.Helper()
	if wantMsg == "" {
		if got != nil {
			t.Fatalf("%s: expected no error, got: %v", msg, got)
		}
		return
	}

	if got == nil {
		t.Fatalf("%s: expected error '%s', got nil", msg, wantMsg)
	}

	gotMsg := got.Error()
	if gotMsg != wantMsg {
		t.Fatalf("%s: expected error '%s', got: '%s'", msg, wantMsg, gotMsg)
	}
}

func NewDomainEvent(entityName commands.AggregateType, aggregateID string, action commands.Action, payload any) *commands.AnyMessage {
	return commands.ToAny(commands.NewDomainEvent(action, payload, entityName, aggregateID, (*commands.AnyMessage)(nil)))
}

type ReducerTestCase[T any] struct {
	Name     string
	Initial  T
	Event    *commands.AnyMessage
	Expected T
}

func RunReducerTests[T any](t *testing.T, tests []ReducerTestCase[T], reducer func(T, *commands.AnyMessage) T) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := reducer(tt.Initial, tt.Event)
			AssertEqual(t, result, tt.Expected, "state after reduction")
		})
	}
}

func RunReducerTestsWithCustomAssert[T any](
	t *testing.T,
	tests []ReducerTestCase[T],
	reducer func(T, *commands.AnyMessage) T,
	assertFunc func(*testing.T, T, T),
) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := reducer(tt.Initial, tt.Event)
			assertFunc(t, result, tt.Expected)
		})
	}
}
