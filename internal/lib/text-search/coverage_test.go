package textsearch

import (
	test_helpers "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestCalculateCoverage(t *testing.T) {
	tests := []struct {
		Name     string
		Text     string
		Input    []string
		Expected float64
	}{
		{
			Name:     "No subtexts returns 0.0",
			Text:     "Climate change impacts global warming.",
			Input:    []string{},
			Expected: 0.0,
		},
		{
			Name:     "Single subtext coverage",
			Text:     "Climate change impacts global warming.",
			Input:    []string{"Climate change"},
			Expected: 0.37,
		},
		{
			Name:     "Multiple subtexts coverage",
			Text:     "Climate change impacts global warming.",
			Input:    []string{"Climate change", "global warming"},
			Expected: 0.74,
		},
		{
			Name:     "All text covered",
			Text:     "Climate change impacts global warming.",
			Input:    []string{"Climate change impacts global warming."},
			Expected: 1.0,
		},
		{
			Name:     "Empty text returns 0.0",
			Text:     "",
			Input:    []string{"test"},
			Expected: 0.0,
		},
		{
			Name:     "Overlapping subtexts can exceed 1.0",
			Text:     "Short text",
			Input:    []string{"Short text", "text", "Short"},
			Expected: 1.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			result := CalculateCoverage(tt.Text, tt.Input)
			if result != tt.Expected {
				t.Errorf("CalculateCoverage() = %v, expected %v", result, tt.Expected)
			}
		})
	}
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
