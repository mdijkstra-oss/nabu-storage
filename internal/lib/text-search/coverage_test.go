package textsearch

import (
	test_helpers "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestCalculateCoverage(t *testing.T) {
	text := "Climate change impacts global warming."

	tests := []struct {
		Name     string
		Input    []string
		Expected float64
	}{
		{
			Name:     "No subtexts returns 0.0",
			Input:    []string{},
			Expected: 0.0,
		},
		{
			Name:     "Single subtext coverage",
			Input:    []string{"Climate change"},
			Expected: 0.37,
		},
		{
			Name:     "Multiple subtexts coverage",
			Input:    []string{"Climate change", "global warming"},
			Expected: 0.74,
		},
		{
			Name:     "All text covered",
			Input:    []string{"Climate change impacts global warming."},
			Expected: 1.0,
		},
	}

	test_helpers.RunFunctionTests(t, tests, func(subTexts []string) float64 {
		return CalculateCoverage(text, subTexts)
	})

	t.Run("Empty text returns 0.0", func(t *testing.T) {
		result := CalculateCoverage("", []string{"test"})
		if result != 0.0 {
			t.Errorf("CalculateCoverage() = %v, expected 0.0", result)
		}
	})

	t.Run("Overlapping subtexts can exceed 1.0", func(t *testing.T) {
		result := CalculateCoverage("Short text", []string{"Short text", "text", "Short"})
		if result != 1.9 {
			t.Errorf("CalculateCoverage() = %v, expected 1.9", result)
		}
	})
}

func TestContainsText(t *testing.T) {
	text := "Climate change impacts global warming."

	tests := []struct {
		Name     string
		Input    string
		Expected bool
	}{
		{
			Name:     "Exact match",
			Input:    "Climate change",
			Expected: true,
		},
		{
			Name:     "Case insensitive",
			Input:    "CLIMATE",
			Expected: true,
		},
		{
			Name:     "No match",
			Input:    "quantum physics",
			Expected: false,
		},
		{
			Name:     "Empty search text",
			Input:    "",
			Expected: false,
		},
	}

	test_helpers.RunFunctionTests(t, tests, func(searchText string) bool {
		return ContainsText(text, searchText)
	})
}
