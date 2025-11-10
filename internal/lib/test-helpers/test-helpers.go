package test_helpers

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"hermes-relay/internal/lib/utils"
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

func RunFunctionTests[T any, R any, M any](t *testing.T, tests []struct {
	Name     string
	Input    T
	Expected M
}, testFunc func(T) R, mapFunc ...func(R) M) {
	testsWithErr := utils.Map(tests, func(tt struct {
		Name     string
		Input    T
		Expected M
	}) struct {
		Name      string
		Input     T
		Expected  M
		ExpectErr string
	} {
		return struct {
			Name      string
			Input     T
			Expected  M
			ExpectErr string
		}{
			Name:      tt.Name,
			Input:     tt.Input,
			Expected:  tt.Expected,
			ExpectErr: "",
		}
	})

	testFuncWithErr := func(input T) (R, error) {
		return testFunc(input), nil
	}

	RunFunctionTestsWithError(t, testsWithErr, testFuncWithErr, mapFunc...)
}

func RunFunctionTestsWithError[T any, R any, M any](t *testing.T, tests []struct {
	Name      string
	Input     T
	Expected  M
	ExpectErr string
}, testFunc func(T) (R, error), mapFunc ...func(R) M) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result, err := testFunc(tt.Input)

			AssertError(t, err, tt.ExpectErr, "error")
			if tt.ExpectErr == "" {
				var actual any
				if len(mapFunc) > 0 {
					actual = mapFunc[0](result)
				} else {
					actual = result
				}
				AssertEqual(t, actual, tt.Expected, "result")
			}
		})
	}
}
