package test_helpers

import (
	"testing"
)

// AssertEqualSimple asserts that two values are equal using simple == comparison
func AssertEqualSimple(t *testing.T, expected, actual any) {
	t.Helper()
	if expected != actual {
		t.Errorf("expected %v, got %v", expected, actual)
	}
}

// AssertNotEmpty asserts that a string is not empty
func AssertNotEmpty(t *testing.T, s string) {
	t.Helper()
	if s == "" {
		t.Error("expected non-empty string")
	}
}

// AssertContains asserts that a string contains a substring
func AssertContains(t *testing.T, s, substr string) {
	t.Helper()
	if substr != "" && len(s) > 0 {
		found := false
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected string to contain %q, got %q", substr, s)
		}
	}
}
