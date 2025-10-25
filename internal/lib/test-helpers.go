package th

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"reflect"
	"testing"
)

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

// AssertEqualIgnoreFields compares two values but ignores specified fields in the given struct type
// Usage: th.AssertEqualIgnoreFields(t, got, want, "msg", file.Chunk{}, "ID")
func AssertEqualIgnoreFields(t *testing.T, got, want any, msg string, ignoreType any, fields ...string) {
	t.Helper()
	AssertEqual(t, got, want, msg, cmpopts.IgnoreFields(ignoreType, fields...))
}
