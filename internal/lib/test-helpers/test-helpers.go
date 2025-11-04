package test_helpers

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"hermes-relay/internal/cqrs"
	"reflect"
	"testing"
)

func AssertMessage(t *testing.T, got, want *cqrs.AnyMessage, msg string) {
	t.Helper()
	AssertEqualIgnoreFields(t, got, want, msg, cqrs.AnyMessage{}, "ID", "Timestamp", "CausationID")
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

func AssertNil(t *testing.T, got any, msg string) {
	t.Helper()
	gotVal := reflect.ValueOf(got)
	if gotVal.IsValid() && !(gotVal.Kind() == reflect.Ptr && gotVal.IsNil()) {
		t.Fatalf("%s: expected nil, got %v", msg, got)
	}
}

func AssertNotNil(t *testing.T, got any, msg string) {
	t.Helper()
	gotVal := reflect.ValueOf(got)
	if !gotVal.IsValid() || (gotVal.Kind() == reflect.Ptr && gotVal.IsNil()) {
		t.Fatalf("%s: expected non-nil, got nil", msg)
	}
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
