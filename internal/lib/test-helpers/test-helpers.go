package test_helpers

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func AssertEqual(t *testing.T, got, want any, msg string, opts ...cmp.Option) {
	t.Helper()

	if isNil(got) && isNil(want) {
		return
	}

	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Fatalf("%s: mismatch (-want +got):\n%s", msg, diff)
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
	if !strings.Contains(gotMsg, wantMsg) {
		t.Fatalf("%s: expected error containing '%s', got: '%s'", msg, wantMsg, gotMsg)
	}
}

func isNil(v any) bool {
	val := reflect.ValueOf(v)
	return !val.IsValid() || (val.Kind() == reflect.Ptr && val.IsNil())
}
