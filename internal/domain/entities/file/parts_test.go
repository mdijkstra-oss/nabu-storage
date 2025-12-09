package file

import (
	"hermes-relay/internal/lib/utils"
	"strings"
	"testing"
)

func TestSplitIntoParts(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		codes    []CodedSection
		maxSize  int
		expected []FilePart
	}{
		{
			name:     "Empty content returns empty parts",
			content:  "",
			codes:    []CodedSection{},
			maxSize:  1000,
			expected: []FilePart{},
		},
		{
			name:    "Content smaller than maxSize returns single part",
			content: "Short content",
			codes:   []CodedSection{},
			maxSize: 1000,
			expected: []FilePart{
				{Content: "Short content", Codes: []CodedSection{}},
			},
		},
		{
			name:    "Content with no newlines splits at maxSize",
			content: strings.Repeat("a", 150),
			codes:   []CodedSection{},
			maxSize: 100,
			expected: []FilePart{
				{Content: strings.Repeat("a", 100), Codes: []CodedSection{}},
				{Content: strings.Repeat("a", 50), Codes: []CodedSection{}},
			},
		},
		{
			name:    "Content splits at newline boundaries",
			content: "First line\nSecond line\nThird line\n",
			codes:   []CodedSection{},
			maxSize: 10,
			expected: []FilePart{
				{Content: "First line\n", Codes: []CodedSection{}},
				{Content: "Second line\n", Codes: []CodedSection{}},
				{Content: "Third line\n", Codes: []CodedSection{}},
			},
		},
		{
			name:    "Codes assigned to correct parts by position",
			content: "First section here\nSecond section here\nThird section here\n",
			codes: []CodedSection{
				{ID: "c1", CodeID: "code-1", Text: "First section", Confidence: ConfidenceHigh},
				{ID: "c2", CodeID: "code-2", Text: "Second section", Confidence: ConfidenceMedium},
				{ID: "c3", CodeID: "code-3", Text: "Third section", Confidence: ConfidenceLow},
			},
			maxSize: 18,
			expected: []FilePart{
				{
					Content: "First section here\n",
					Codes: []CodedSection{
						{ID: "c1", CodeID: "code-1", Text: "First section", Confidence: ConfidenceHigh},
					},
				},
				{
					Content: "Second section here\n",
					Codes: []CodedSection{
						{ID: "c2", CodeID: "code-2", Text: "Second section", Confidence: ConfidenceMedium},
					},
				},
				{
					Content: "Third section here\n",
					Codes: []CodedSection{
						{ID: "c3", CodeID: "code-3", Text: "Third section", Confidence: ConfidenceLow},
					},
				},
			},
		},
		{
			name:    "Multiple codes in same part",
			content: "This has two codes here and another code there\n",
			codes: []CodedSection{
				{ID: "c1", CodeID: "code-1", Text: "two codes", Confidence: ConfidenceHigh},
				{ID: "c2", CodeID: "code-2", Text: "another code", Confidence: ConfidenceMedium},
			},
			maxSize: 100,
			expected: []FilePart{
				{
					Content: "This has two codes here and another code there\n",
					Codes: []CodedSection{
						{ID: "c1", CodeID: "code-1", Text: "two codes", Confidence: ConfidenceHigh},
						{ID: "c2", CodeID: "code-2", Text: "another code", Confidence: ConfidenceMedium},
					},
				},
			},
		},
		{
			name:    "Code not found in content excluded from all parts",
			content: "Some content here\n",
			codes: []CodedSection{
				{ID: "c1", CodeID: "code-1", Text: "nonexistent text", Confidence: ConfidenceHigh},
			},
			maxSize: 100,
			expected: []FilePart{
				{Content: "Some content here\n", Codes: []CodedSection{}},
			},
		},
		{
			name:    "Large content with realistic maxSize",
			content: strings.Repeat("Line of text\n", 1000),
			codes: []CodedSection{
				{ID: "c1", CodeID: "code-1", Text: "Line of text", Confidence: ConfidenceHigh},
			},
			maxSize: 12000,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SplitIntoParts(tt.content, tt.codes, tt.maxSize)

			if tt.expected == nil {
				if len(result) == 0 {
					t.Fatal("Expected multiple parts for large content, got 0")
				}

				totalContent := ""
				for _, part := range result {
					totalContent += part.Content
				}
				if totalContent != tt.content {
					t.Error("Reconstructed content does not match original")
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d parts, got %d", len(tt.expected), len(result))
			}

			for i, part := range result {
				if part.Content != tt.expected[i].Content {
					t.Errorf("Part %d content mismatch:\nExpected: %q\nGot: %q",
						i, tt.expected[i].Content, part.Content)
				}

				if len(part.Codes) != len(tt.expected[i].Codes) {
					t.Errorf("Part %d expected %d codes, got %d",
						i, len(tt.expected[i].Codes), len(part.Codes))
					continue
				}

				for j, code := range part.Codes {
					if code.ID != tt.expected[i].Codes[j].ID {
						t.Errorf("Part %d code %d ID mismatch: expected %s, got %s",
							i, j, tt.expected[i].Codes[j].ID, code.ID)
					}
				}
			}
		})
	}
}

func TestFindNearestNewline(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		position int
		expected int
	}{
		{
			name:     "Position at newline returns next position",
			content:  "abc\ndef",
			position: 3,
			expected: 4,
		},
		{
			name:     "Position after newline finds next newline",
			content:  "abc\ndef\nghi",
			position: 5,
			expected: 8,
		},
		{
			name:     "Position before newline finds nearest forward",
			content:  "abc\ndef",
			position: 2,
			expected: 4,
		},
		{
			name:     "No newline after position finds previous",
			content:  "abc\ndef",
			position: 6,
			expected: 4,
		},
		{
			name:     "No newlines returns position",
			content:  "abcdef",
			position: 3,
			expected: 3,
		},
		{
			name:     "Position beyond content returns content length",
			content:  "abc",
			position: 10,
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findNearestNewline(tt.content, tt.position)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestFindCodePositions(t *testing.T) {
	tests := []struct {
		name     string
		codes    []CodedSection
		content  string
		expected map[string]int
	}{
		{
			name:     "Empty codes returns empty map",
			codes:    []CodedSection{},
			content:  "Some content",
			expected: map[string]int{},
		},
		{
			name: "Single code found at position",
			codes: []CodedSection{
				{ID: "c1", Text: "found text"},
			},
			content:  "This is found text here",
			expected: map[string]int{"c1": 8},
		},
		{
			name: "Multiple codes at different positions",
			codes: []CodedSection{
				{ID: "c1", Text: "first"},
				{ID: "c2", Text: "second"},
			},
			content:  "The first and second words",
			expected: map[string]int{"c1": 4, "c2": 14},
		},
		{
			name: "Code not found excluded from map",
			codes: []CodedSection{
				{ID: "c1", Text: "exists"},
				{ID: "c2", Text: "missing"},
			},
			content:  "This exists here",
			expected: map[string]int{"c1": 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findCodePositions(tt.codes, tt.content)

			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d positions, got %d", len(tt.expected), len(result))
			}

			for id, expectedPos := range tt.expected {
				if pos, found := result[id]; !found {
					t.Errorf("Expected position for code %s, but not found", id)
				} else if pos != expectedPos {
					t.Errorf("Code %s: expected position %d, got %d", id, expectedPos, pos)
				}
			}
		})
	}
}

func TestCodesInRange(t *testing.T) {
	codes := []CodedSection{
		{ID: "c1", Text: "first"},
		{ID: "c2", Text: "second"},
		{ID: "c3", Text: "third"},
	}

	positions := map[string]int{
		"c1": 10,
		"c2": 50,
		"c3": 90,
	}

	tests := []struct {
		name     string
		start    int
		end      int
		expected []string
	}{
		{
			name:     "Range includes first code only",
			start:    0,
			end:      40,
			expected: []string{"c1"},
		},
		{
			name:     "Range includes middle code only",
			start:    40,
			end:      80,
			expected: []string{"c2"},
		},
		{
			name:     "Range includes all codes",
			start:    0,
			end:      100,
			expected: []string{"c1", "c2", "c3"},
		},
		{
			name:     "Range includes no codes",
			start:    0,
			end:      5,
			expected: []string{},
		},
		{
			name:     "Range boundary excludes code at exact end",
			start:    0,
			end:      50,
			expected: []string{"c1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := codesInRange(codes, positions, tt.start, tt.end)

			if len(result) != len(tt.expected) {
				t.Fatalf("Expected %d codes, got %d", len(tt.expected), len(result))
			}

			resultIDs := utils.Map(result, func(c CodedSection) string { return c.ID })
			for _, expectedID := range tt.expected {
				if !utils.Contains(resultIDs, expectedID) {
					t.Errorf("Expected code %s in result", expectedID)
				}
			}
		})
	}
}
