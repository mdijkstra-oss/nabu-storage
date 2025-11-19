package find

import (
	"testing"
)

func TestFind_Comprehensive(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		needle       string
		expectFound  bool
		expectActual string // What we expect to actually extract from text
	}{
		{
			name:         "simple sentence",
			text:         "The quick brown fox jumps over the lazy dog.",
			needle:       "quick brown fox",
			expectFound:  true,
			expectActual: "quick brown fox",
		},
		{
			name:         "sentence with apostrophe",
			text:         "John's dog is very friendly and loves to play.",
			needle:       "John's dog is very friendly",
			expectFound:  true,
			expectActual: "John's dog is very friendly",
		},
		{
			name:         "apostrophe fuzzy match - john's matches johns",
			text:         "Johns dog ran quickly down the street.",
			needle:       "John's dog ran quickly",
			expectFound:  true,
			expectActual: "Johns dog ran quickly", // Apostrophe stripped in source
		},
		{
			name:         "sentence with em dash",
			text:         "The forest was quiet—almost unnaturally so.",
			needle:       "forest was quiet—almost unnaturally",
			expectFound:  true,
			expectActual: "forest was quiet—almost unnaturally",
		},
		{
			name:         "em dash creates word boundary",
			text:         "Everything was balanced—the deer, the stream.",
			needle:       "balanced—the deer",
			expectFound:  true,
			expectActual: "balanced—the deer,",
		},
		{
			name:         "text with quotes in needle",
			text:         "She said \"this is important\" to everyone.",
			needle:       "\"this is important\"",
			expectFound:  true,
			expectActual: "\"this is important\"",
		},
		{
			name:         "tab separated",
			text:         "First column\tSecond column\tThird column",
			needle:       "Second column\tThird",
			expectFound:  true,
			expectActual: "Second column\tThird",
		},
		{
			name:         "newlines in text",
			text:         "First line\nSecond line\nThird line",
			needle:       "Second line\nThird",
			expectFound:  true,
			expectActual: "Second line\nThird",
		},
		{
			name:         "hyphenated words",
			text:         "This is a well-known fact about user-friendly design.",
			needle:       "well-known fact about user-friendly",
			expectFound:  true,
			expectActual: "well-known fact about user-friendly",
		},
		{
			name:         "comma separated list",
			text:         "We need apples, oranges, bananas, and grapes.",
			needle:       "oranges, bananas, and grapes",
			expectFound:  true,
			expectActual: "oranges, bananas, and grapes.",
		},
		{
			name:         "parentheses in text",
			text:         "This is important (really important) for understanding.",
			needle:       "important (really important) for",
			expectFound:  true,
			expectActual: "important (really important) for",
		},
		{
			name:         "colon and semicolon",
			text:         "Remember this: always be kind; treat others well.",
			needle:       "always be kind; treat others",
			expectFound:  true,
			expectActual: "always be kind; treat others",
		},
		{
			name:         "multiple punctuation marks",
			text:         "Really?! Yes, absolutely! This is correct.",
			needle:       "Yes, absolutely! This is",
			expectFound:  true,
			expectActual: "Yes, absolutely! This is",
		},
		{
			name:         "ellipsis",
			text:         "Well... I'm not sure about this.",
			needle:       "I'm not sure about",
			expectFound:  true,
			expectActual: "I'm not sure about",
		},
		{
			name:         "mixed case normalized",
			text:         "The IMPORTANT thing is CLARITY.",
			needle:       "important thing is clarity",
			expectFound:  true,
			expectActual: "IMPORTANT thing is CLARITY.",
		},
		{
			name:         "trailing punctuation",
			text:         "This is the end...",
			needle:       "This is the end",
			expectFound:  true,
			expectActual: "This is the end...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			foundText, found := Find(tt.needle, tt.text)

			if found != tt.expectFound {
				t.Fatalf("Expected found=%v, got found=%v", tt.expectFound, found)
			}

			if !tt.expectFound {
				return
			}

			t.Logf("Needle: %q", tt.needle)
			t.Logf("Found:  %q", foundText)

			if foundText != tt.expectActual {
				t.Errorf("Text mismatch:\nExpected: %q\nActual:   %q", tt.expectActual, foundText)
			}
		})
	}
}
