package find

import "testing"

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"nbsp to space", "hello\u00A0world", "hello world"},
		{"curly apostrophe removed", "it\u2019s", "its"},
		{"straight apostrophe removed", "it's", "its"},
		{"curly double quotes removed", "\u201Chello\u201D", "hello"},
		{"straight double quotes removed", "\"hello\"", "hello"},
		{"curly single quotes removed", "\u2018hello\u2019", "hello"},
		{"underscore to space", "_hello_", " hello "},
		{"asterisk to space", "*hello*", " hello "},
		{"backtick to space", "`code`", " code "},
		{"mixed markdown", "**bold** and _italic_", "  bold   and  italic "},
		{"unchanged text", "hello world", "hello world"},
		{"multiple nbsp", "a\u00A0\u00A0b", "a  b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeText(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeText(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
