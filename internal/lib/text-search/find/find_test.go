package find

import (
	"os"
	"strings"
	"testing"
)

// =====================================================
// Helper types and functions
// =====================================================

func assertFound(t *testing.T, needle, searchText, foundText string, found bool) {
	t.Helper()
	if !found {
		t.Fatalf("Expected to find %q in %q, but got found=false", needle, searchText)
	}
	if foundText == "" {
		t.Fatalf("Expected non-empty found text, but got empty string")
	}
}

func assertNotFound(t *testing.T, needle, searchText, foundText string, found bool) {
	t.Helper()
	if found {
		t.Errorf("Expected not to find %q, but got found=true with text %q", needle, foundText)
	}
	if foundText != "" {
		t.Errorf("Expected empty text for not found, got %q", foundText)
	}
}

// =====================================================
// Unified test runner
// =====================================================

type FindRangeTestCase struct {
	name     string
	text     string
	needle   string
	wantFind bool
}

func runFindRangeTests(t *testing.T, tests []FindRangeTestCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundText, found := FindRange(tt.needle, tt.text)

			if tt.wantFind {
				assertFound(t, tt.needle, tt.text, foundText, found)
				// Verify that the found text actually exists in the original text
				if !strings.Contains(tt.text, foundText) {
					t.Errorf("Found text %q does not exist in original text", foundText)
				}
			} else {
				assertNotFound(t, tt.needle, tt.text, foundText, found)
			}
		})
	}
}

// =====================================================
// Tests
// =====================================================

func TestFindRange_FuzzyMatching(t *testing.T) {
	tests := []FindRangeTestCase{
		// Exact matches
		{"exact match", "The quick brown fox jumps over the lazy dog", "brown fox jumps", true},
		{"single word", "one two three four", "three", true},
		{"entire text", "This is the complete text", "This is the complete text", true},

		// Extra/missing spaces
		{"extra spaces in text", "The  dog    ran  fast", "dog ran fast", true},
		{"extra spaces in needle", "The dog ran fast", "dog  ran  fast", true},
		{"extra spaces both", "The  dog  ran  fast", "dog   ran   fast", true},

		// Newline differences
		{"newline in text", "line one\nline two\nline three", "line one line two", true},
		{"newline in needle", "line one line two line three", "line one\nline two", true},
		{"multiple newlines", "first\n\nsecond\n\nthird", "first second third", true},

		// Punctuation variations
		{"comma in text", "Hello, world! How are you?", "Hello world How are", true},
		{"punctuation removed", "Well... this is interesting!", "Well this is interesting", true},
		{"mixed punctuation", "The team's goal: ship fast, ship well.", "teams goal ship fast ship well", true},

		// Boundary conditions
		{"start of text", "Start here with some text following after", "Start here with", true},
		{"end of text", "Some text in the middle ending here", "ending here", true},

		// Case insensitive
		{"different case", "Hello World", "hello world", true},

		// Not found
		{"non-existent", "The quick brown fox", "purple elephant dancing", false},
		{"empty needle", "Some text here", "", false},
		{"empty text", "", "something", false},
		{"needle too long", "short", "this is a much longer needle than the text", false},
	}

	runFindRangeTests(t, tests)
}

func TestFindRange_RealisticDocument(t *testing.T) {
	content, err := os.ReadFile("realistic-document.md")
	if err != nil {
		t.Fatalf("Failed to read realistic document: %v", err)
	}
	text := string(content)

	tests := []FindRangeTestCase{
		{"heading text", text, "User Research Interview Notes", true},
		{"paragraph with punctuation", text, "The main challenge is context switching", true},
		{"list item", text, "Too many disconnected tools", true},
		{"quoted text", text, "We spend 60% of our time just trying to remember what we learned yesterday", true},
		{"text with extra spaces", text, "Sarah mentioned that she often loses track", true},
		{"numbered list item", text, "Collect raw data from multiple sources", true},
		{"conclusion paragraph", text, "The key insight is that researchers need tools that preserve context", true},
		{"technical terms", text, "AI-assisted summarization", true},
		{"team needs", text, "Sarah's team needs a solution", true},
	}

	runFindRangeTests(t, tests)
}

func TestFindRange_EdgeCases(t *testing.T) {
	tests := []FindRangeTestCase{
		{"empty needle", "Some text here", "", false},
		{"empty text", "", "something", false},
		{"both empty", "", "", false},
		{"needle longer than text", "short", "this is much longer", false},
		{"single character text", "a", "a", true},
		{"single character needle", "abc def ghi", "def", true},
		{"whitespace only text", "   \n\t  ", "test", false},
		{"whitespace only needle", "some text", "   ", false},
	}

	runFindRangeTests(t, tests)
}
