package find

import "testing"

func TestBalanceMarkdownTags(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		selection string
		expected  string
	}{
		{"already balanced", "hello **world** there", "**world**", "**world**"},
		{"unbalanced bold - expand end", "the **targeted lockdown** is here", "**targeted", "**targeted lockdown**"},
		{"unbalanced bold - expand start", "the **targeted lockdown** is here", "targeted lockdown**", "**targeted lockdown**"},
		{"unbalanced italic underscore", "the _targeted lockdown_ is here", "_targeted", "_targeted lockdown_"},
		{"unbalanced backtick", "use `code here` please", "`code", "`code here`"},
		{"no markdown", "plain text here", "text", "text"},
		{"nested bold and italic", "the **_nested_** here", "**_nes", "**_nested_**"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := findSubstring(tt.text, tt.selection)
			if start == -1 {
				t.Fatalf("selection %q not found in text %q", tt.selection, tt.text)
			}

			gotStart, gotEnd := BalanceMarkdownTags(tt.text, start, end)
			got := tt.text[gotStart:gotEnd]

			if got != tt.expected {
				t.Errorf("BalanceMarkdownTags(%q, %q) = %q, want %q",
					tt.text, tt.selection, got, tt.expected)
			}
		})
	}
}

func findSubstring(text, sub string) (int, int) {
	for i := 0; i <= len(text)-len(sub); i++ {
		if text[i:i+len(sub)] == sub {
			return i, i + len(sub)
		}
	}
	return -1, -1
}
