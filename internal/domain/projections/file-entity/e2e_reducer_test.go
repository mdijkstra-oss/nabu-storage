package fileview

import (
	"bufio"
	"encoding/json"
	"hermes-relay/internal/cqrs/commands"
	"os"
	"path/filepath"
	"testing"
)

type E2EReducerTestCase struct {
	Name          string
	EventsFile    string
	ValidateState func(t *testing.T, final *File)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

func replayEvents(events []*commands.AnyMessage, reducer func(*File, *commands.AnyMessage) *File) *File {
	var state *File
	for _, event := range events {
		state = reducer(state, event)
	}
	return state
}

func validateCodedSectionTexts(t *testing.T, file *File) {
	t.Helper()

	for _, chunk := range file.Chunks {
		for _, code := range chunk.Codes {
			if code.Text == "" {
				t.Errorf("Empty text for coded section %s in chunk %s", code.CodeSlug, chunk.ID)
				continue
			}

			// Verify the text actually exists in the chunk content
			if !containsText(chunk.Content, code.Text) {
				t.Errorf(
					"Coded section text not found in chunk %s, code %s\n  Text: %q",
					chunk.ID, code.CodeSlug, code.Text,
				)
			}
		}
	}
}

func containsText(content, text string) bool {
	// Simple substring search - the text should be an exact substring
	for i := 0; i <= len(content)-len(text); i++ {
		if content[i:i+len(text)] == text {
			return true
		}
	}
	return false
}

func TestE2EReducer(t *testing.T) {
	tests := []E2EReducerTestCase{
		{
			Name:       "Broken test should fail",
			EventsFile: "testdata/broken-test.ndjson",
			ValidateState: func(t *testing.T, final *File) {
				if final == nil {
					t.Fatal("Expected non-nil file state")
				}
				chunk := final.Chunks[0]
				if len(chunk.Codes) != 7 {
					t.Fatalf("Expected 7 coded sections (9 minus 2 broken ones), got %d", len(chunk.Codes))
				}
				validateCodedSectionTexts(t, final)
			},
		},
		{
			Name:       "Nature test with multiple coded sections",
			EventsFile: "testdata/nature-test.ndjson",
			ValidateState: func(t *testing.T, final *File) {
				if final == nil {
					t.Fatal("Expected non-nil file state")
				}
				if final.Name != "Morning by the Stream" {
					t.Errorf("Expected name 'Morning by the Stream', got %q", final.Name)
				}
				if len(final.Chunks) != 1 {
					t.Fatalf("Expected 1 chunk, got %d", len(final.Chunks))
				}

				chunk := final.Chunks[0]
				expectedCodeCount := 9
				if len(chunk.Codes) != expectedCodeCount {
					t.Logf("Actual coded sections found:")
					for i, code := range chunk.Codes {
						t.Logf("  [%d] %s: %q", i, code.CodeSlug, code.Text[:min(50, len(code.Text))])
					}
					t.Fatalf("Expected %d coded sections, got %d", expectedCodeCount, len(chunk.Codes))
				}

				expectedCodes := map[string]int{
					"setting:forest":    2,
					"element:water":     3,
					"element:wildlife":  1,
					"mood:tranquil":     3,
				}

				actualCodes := make(map[string]int)
				for _, code := range chunk.Codes {
					actualCodes[code.CodeSlug]++
				}

				for slug, expectedCount := range expectedCodes {
					if actualCodes[slug] != expectedCount {
						t.Errorf("Expected %d instances of %q, got %d", expectedCount, slug, actualCodes[slug])
					}
				}

				validateCodedSectionTexts(t, final)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			eventsPath := filepath.Join(tt.EventsFile)
			events := loadEventsFromFile(t, eventsPath)

			if len(events) == 0 {
				t.Fatal("No events loaded from file")
			}

			final := replayEvents(events, Reducer)
			tt.ValidateState(t, final)
		})
	}
}
