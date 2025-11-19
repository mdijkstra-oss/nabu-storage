package fileview

import (
	"bufio"
	"encoding/json"
	"hermes-relay/internal/cqrs/commands"
	"os"
	"reflect"
	"testing"
)

type E2EReducerTestCase struct {
	Name         string
	EventsFile   string
	ExpectedFile string
}

func loadEventsFromFile(t *testing.T, filePath string) []*commands.AnyMessage {
	t.Helper()

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open events file %s: %v", filePath, err)
	}
	defer file.Close()

	var events []*commands.AnyMessage
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var msg commands.AnyMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			t.Fatalf("Failed to unmarshal event: %v", err)
		}
		events = append(events, &msg)
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading events file: %v", err)
	}

	return events
}

func loadExpectedFromFile(t *testing.T, filePath string) *File {
	t.Helper()

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read expected file %s: %v", filePath, err)
	}

	var expected File
	if err := json.Unmarshal(data, &expected); err != nil {
		t.Fatalf("Failed to unmarshal expected file: %v", err)
	}

	return &expected
}

func replayEvents(events []*commands.AnyMessage, reducer func(*File, *commands.AnyMessage) *File) *File {
	var state *File
	for _, event := range events {
		state = reducer(state, event)
	}
	return state
}

func compareFiles(t *testing.T, actual, expected *File) {
	t.Helper()

	// Normalize time field for comparison (reducer sets to time.Now())
	actualCopy := *actual
	actualCopy.Time = expected.Time

	if !reflect.DeepEqual(&actualCopy, expected) {
		actualJSON, _ := json.MarshalIndent(&actualCopy, "", "  ")
		expectedJSON, _ := json.MarshalIndent(expected, "", "  ")
		t.Errorf("File mismatch:\n\nActual:\n%s\n\nExpected:\n%s", actualJSON, expectedJSON)
	}
}

func TestE2EReducer(t *testing.T) {
	tests := []E2EReducerTestCase{
		{
			Name:         "Fuzzy matching with spacing issues",
			EventsFile:   "testdata/spacing_fuzzy.jsonl",
			ExpectedFile: "testdata/spacing_fuzzy.expected.json",
		},
		{
			Name:         "Wrong text should not be found",
			EventsFile:   "testdata/not_found.jsonl",
			ExpectedFile: "testdata/not_found.expected.json",
		},
		{
			Name:         "Punctuation and newline variations",
			EventsFile:   "testdata/punctuation_newlines.jsonl",
			ExpectedFile: "testdata/punctuation_newlines.expected.json",
		},
		{
			Name:         "Casing and punctuation differences",
			EventsFile:   "testdata/casing_punctuation.jsonl",
			ExpectedFile: "testdata/casing_punctuation.expected.json",
		},
		{
			Name:         "Minimum 3 words required",
			EventsFile:   "testdata/min_words.jsonl",
			ExpectedFile: "testdata/min_words.expected.json",
		},
		{
			Name:         "Edge cases: whitespace and hyphens",
			EventsFile:   "testdata/edge_cases.jsonl",
			ExpectedFile: "testdata/edge_cases.expected.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			events := loadEventsFromFile(t, tt.EventsFile)
			expected := loadExpectedFromFile(t, tt.ExpectedFile)

			if len(events) == 0 {
				t.Fatal("No events loaded from file")
			}

			actual := replayEvents(events, Reducer)
			compareFiles(t, actual, expected)
		})
	}
}
