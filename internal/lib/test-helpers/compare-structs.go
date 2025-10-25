package th

import (
	"reflect"
	"testing"
)

// AssertStructEquality uses reflection to compare two structs and fatally fails if different
func AssertStructEquality(t *testing.T, got, want any, msg string) {
	t.Helper()

	gotVal := reflect.ValueOf(got)
	wantVal := reflect.ValueOf(want)

	// Handle nil cases
	gotIsNil := !gotVal.IsValid() || (gotVal.Kind() == reflect.Ptr && gotVal.IsNil())
	wantIsNil := !wantVal.IsValid() || (wantVal.Kind() == reflect.Ptr && wantVal.IsNil())

	if gotIsNil && wantIsNil {
		return
	}
	if gotIsNil {
		t.Fatalf("%s: got nil, want %+v", msg, want)
	}
	if wantIsNil {
		t.Fatalf("%s: got %+v, want nil", msg, got)
	}

	// Dereference pointers
	if gotVal.Kind() == reflect.Ptr {
		gotVal = gotVal.Elem()
	}
	if wantVal.Kind() == reflect.Ptr {
		wantVal = wantVal.Elem()
	}

	if gotVal.Type() != wantVal.Type() {
		t.Fatalf("%s: type mismatch: got %v, want %v", msg, gotVal.Type(), wantVal.Type())
	}

	// Compare struct fields
	for i := 0; i < gotVal.NumField(); i++ {
		fieldName := gotVal.Type().Field(i).Name
		gotField := gotVal.Field(i)
		wantField := wantVal.Field(i)

		if !reflect.DeepEqual(gotField.Interface(), wantField.Interface()) {
			t.Fatalf("%s: field %s: got %v, want %v",
				msg,
				fieldName,
				gotField.Interface(),
				wantField.Interface())
		}
	}
}
