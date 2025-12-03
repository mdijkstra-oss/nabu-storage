package find

import (
	"os"
	"strings"
	"testing"
)

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

type FindTestCase struct {
	name         string
	text         string
	needle       string
	wantFind     bool
	expectActual string
}

func runFindTests(t *testing.T, tests []FindTestCase) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundText, found := Find(tt.needle, tt.text)

			if tt.wantFind {
				assertFound(t, tt.needle, tt.text, foundText, found)
				if !strings.Contains(tt.text, foundText) {
					t.Errorf("Found text %q does not exist in original text", foundText)
				}
				if tt.expectActual != "" && foundText != tt.expectActual {
					t.Errorf("Text mismatch:\nExpected: %q\nActual:   %q", tt.expectActual, foundText)
				}
			} else {
				assertNotFound(t, tt.needle, tt.text, foundText, found)
			}
		})
	}
}

func TestFind_FuzzyMatching(t *testing.T) {
	tests := []FindTestCase{
		{"exact match", "The quick brown fox jumps over the lazy dog", "brown fox jumps", true, ""},
		{"single word", "one two three four", "three", true, ""},
		{"entire text", "This is the complete text", "This is the complete text", true, ""},

		{"extra spaces in text", "The  dog    ran  fast", "dog ran fast", true, ""},
		{"extra spaces in needle", "The dog ran fast", "dog  ran  fast", true, ""},
		{"extra spaces both", "The  dog  ran  fast", "dog   ran   fast", true, ""},

		{"newline in text", "line one\nline two\nline three", "line one line two", true, ""},
		{"newline in needle", "line one line two line three", "line one\nline two", true, ""},
		{"multiple newlines", "first\n\nsecond\n\nthird", "first second third", true, ""},

		{"comma in text", "Hello, world! How are you?", "Hello world How are", true, ""},
		{"punctuation removed", "Well... this is interesting!", "Well this is interesting", true, ""},
		{"mixed punctuation", "The team's goal: ship fast, ship well.", "teams goal ship fast ship well", true, ""},

		{"start of text", "Start here with some text following after", "Start here with", true, ""},
		{"end of text", "Some text in the middle ending here", "ending here", true, ""},

		{"different case", "Hello World", "hello world", true, ""},

		{"non-existent", "The quick brown fox", "purple elephant dancing", false, ""},
		{"scrambled word order", "The quick brown fox jumps", "fox brown quick", false, ""},
		{"empty needle", "Some text here", "", false, ""},
		{"empty text", "", "something", false, ""},
		{"needle too long", "short", "this is a much longer needle than the text", false, ""},
	}

	runFindTests(t, tests)
}

func TestFind_RealisticDocument(t *testing.T) {
	content, err := os.ReadFile("realistic-document.md")
	if err != nil {
		t.Fatalf("Failed to read realistic document: %v", err)
	}
	text := string(content)

	tests := []FindTestCase{
		{"heading text", text, "User Research Interview Notes", true, ""},
		{"paragraph with punctuation", text, "The main challenge is context switching", true, ""},
		{"list item", text, "Too many disconnected tools", true, ""},
		{"quoted text", text, "We spend 60% of our time just trying to remember what we learned yesterday", true, ""},
		{"text with extra spaces", text, "Sarah mentioned that she often loses track", true, ""},
		{"numbered list item", text, "Collect raw data from multiple sources", true, ""},
		{"conclusion paragraph", text, "The key insight is that researchers need tools that preserve context", true, ""},
		{"technical terms", text, "AI-assisted summarization", true, ""},
		{"team needs", text, "Sarah's team needs a solution", true, ""},
	}

	runFindTests(t, tests)
}

func TestFind_EdgeCases(t *testing.T) {
	tests := []FindTestCase{
		{"empty needle", "Some text here", "", false, ""},
		{"empty text", "", "something", false, ""},
		{"both empty", "", "", false, ""},
		{"needle longer than text", "short", "this is much longer", false, ""},
		{"single character text", "a", "a", true, ""},
		{"single character needle", "abc def ghi", "def", true, ""},
		{"whitespace only text", "   \n\t  ", "test", false, ""},
		{"whitespace only needle", "some text", "   ", false, ""},
	}

	runFindTests(t, tests)
}

func TestFindAll_MultipleMatches(t *testing.T) {
	matches := FindAll("the cat", "The cat sat. Then the cat ran. Finally the cat slept.")

	if len(matches) != 3 {
		t.Errorf("FindAll() got %d matches, want 3", len(matches))
	}

	wantTexts := []string{"The cat", "the cat", "the cat"}
	for i, m := range matches {
		if m.Text != wantTexts[i] {
			t.Errorf("match[%d].Text = %q, want %q", i, m.Text, wantTexts[i])
		}
	}
}

func TestExtractContext(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		needle  string
		context int
		want    string
	}{
		{"context 1 mid", "First. Match here. Last.", "Match here", 1, "First. Match here. Last."},
		{"context 1 start", "Match here. After.", "Match here", 1, "Match here. After."},
		{"context 1 end", "Before. Match here", "Match here", 1, "Before. Match here"},
		{"context 2 mid", "One. Two. Match. Four. Five.", "Match", 2, "One. Two. Match. Four. Five."},
		{"context 2 not enough before", "Before. Match. Three. Four.", "Match", 2, "Before. Match. Three. Four."},
		{"newline boundary", "Line one\nMatch line\nLine three", "Match line", 1, "Line one\nMatch line\nLine three"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := FindAll(tt.needle, tt.text)
			if len(matches) == 0 {
				t.Fatalf("FindAll(%q, %q) found no matches", tt.needle, tt.text)
			}
			m := matches[0]
			got := ExtractContext(tt.text, m.Start, m.End, tt.context)
			if got != tt.want {
				t.Errorf("ExtractContext() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReplace(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		oldText  string
		newText  string
		expected string
	}{
		{"simple replace", "hello world", "world", "universe", "hello universe"},
		{"replace at start", "foo bar baz", "foo", "qux", "qux bar baz"},
		{"replace at end", "foo bar baz", "baz", "qux", "foo bar qux"},
		{"replace with empty", "hello world", "world", "", "hello "},
		{"replace first occurrence only", "cat cat cat", "cat", "dog", "dog cat cat"},
		{"no match returns unchanged", "hello world", "xyz", "abc", "hello world"},
		{"empty content", "", "foo", "bar", ""},
		{"empty oldText returns unchanged", "hello world", "", "bar", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Replace(tt.content, tt.oldText, tt.newText)
			if result != tt.expected {
				t.Errorf("Replace(%q, %q, %q) = %q, want %q", tt.content, tt.oldText, tt.newText, result, tt.expected)
			}
		})
	}
}
