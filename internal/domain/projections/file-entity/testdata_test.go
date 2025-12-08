package fileview

import (
	"encoding/json"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/persistence"
	test_helpers "hermes-relay/internal/lib/test-helpers"
	"os"
	"path/filepath"
	"testing"
)

type testCase struct {
	name         string
	eventsPath   string
	expectedPath string
}

func testCases() []testCase {
	cases := []string{
		"edge_cases",
		"min_words",
		"casing_punctuation",
		"not_found",
		"spacing_fuzzy",
		"punctuation_newlines",
	}

	result := make([]testCase, len(cases))
	for i, name := range cases {
		result[i] = testCase{
			name:         name,
			eventsPath:   filepath.Join("testdata", name+".jsonl"),
			expectedPath: filepath.Join("testdata", name+".expected.json"),
		}
	}
	return result
}

func TestTextMatchingEdgeCases(t *testing.T) {
	for _, tc := range testCases() {
		t.Run(tc.name, func(t *testing.T) {
			events, err := persistence.ReadEventsFromFile(tc.eventsPath)
			if err != nil {
				t.Fatalf("Failed to load events: %v", err)
			}

			expected, err := loadExpectedFile(tc.expectedPath)
			if err != nil {
				t.Fatalf("Failed to load expected: %v", err)
			}

			actual := reduceEvents(events)

			test_helpers.AssertEqualIgnoreFields(t, actual, expected, "final state", File{}, "Time", "ID", "Version")
		})
	}
}

func reduceEvents(events []commands.AnyMessage) *File {
	var state *File
	for i := range events {
		state = Reducer(state, &events[i])
	}
	return state
}

func loadExpectedFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var expected File
	if err := json.Unmarshal(data, &expected); err != nil {
		return nil, err
	}

	return &expected, nil
}
