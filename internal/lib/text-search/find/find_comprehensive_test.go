package find

import (
	"testing"
)

func TestFind_Comprehensive(t *testing.T) {
	tests := []FindTestCase{
		{"simple sentence", "The quick brown fox jumps over the lazy dog.", "quick brown fox", true, "quick brown fox"},
		{"sentence with apostrophe", "John's dog is very friendly and loves to play.", "John's dog is very friendly", true, "John's dog is very friendly"},
		{"apostrophe fuzzy match - john's matches johns", "Johns dog ran quickly down the street.", "John's dog ran quickly", true, "Johns dog ran quickly"},
		{"sentence with em dash", "The forest was quiet—almost unnaturally so.", "forest was quiet—almost unnaturally", true, "forest was quiet—almost unnaturally"},
		{"em dash creates word boundary", "Everything was balanced—the deer, the stream.", "balanced—the deer", true, "balanced—the deer,"},
		{"text with quotes in needle", "She said \"this is important\" to everyone.", "\"this is important\"", true, "\"this is important\""},
		{"tab separated", "First column\tSecond column\tThird column", "Second column\tThird", true, "Second column\tThird"},
		{"newlines in text", "First line\nSecond line\nThird line", "Second line\nThird", true, "Second line\nThird"},
		{"hyphenated words", "This is a well-known fact about user-friendly design.", "well-known fact about user-friendly", true, "well-known fact about user-friendly"},
		{"comma separated list", "We need apples, oranges, bananas, and grapes.", "oranges, bananas, and grapes", true, "oranges, bananas, and grapes."},
		{"parentheses in text", "This is important (really important) for understanding.", "important (really important) for", true, "important (really important) for"},
		{"colon and semicolon", "Remember this: always be kind; treat others well.", "always be kind; treat others", true, "always be kind; treat others"},
		{"multiple punctuation marks", "Really?! Yes, absolutely! This is correct.", "Yes, absolutely! This is", true, "Yes, absolutely! This is"},
		{"ellipsis", "Well... I'm not sure about this.", "I'm not sure about", true, "I'm not sure about"},
		{"mixed case normalized", "The IMPORTANT thing is CLARITY.", "important thing is clarity", true, "IMPORTANT thing is CLARITY."},
		{"trailing punctuation", "This is the end...", "This is the end", true, "This is the end..."},
	}

	runFindTests(t, tests)
}
