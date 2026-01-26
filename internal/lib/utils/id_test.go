package utils

import "testing"

func TestValidID(t *testing.T) {
	tests := []struct {
		Name     string
		Input    string
		Expected bool
	}{
		{Name: "valid uuid lowercase", Input: "550e8400-e29b-41d4-a716-446655440000", Expected: true},
		{Name: "valid uuid uppercase", Input: "550E8400-E29B-41D4-A716-446655440000", Expected: true},
		{Name: "invalid format", Input: "not-a-uuid", Expected: false},
		{Name: "empty string", Input: "", Expected: false},
		{Name: "uuid without dashes", Input: "550e8400e29b41d4a716446655440000", Expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := ValidID(tt.Input)
			if got != tt.Expected {
				t.Errorf("ValidID(%q) = %v, want %v", tt.Input, got, tt.Expected)
			}
		})
	}
}

func TestValidFilePath(t *testing.T) {
	tests := []struct {
		Name     string
		Input    string
		Expected bool
	}{
		{Name: "simple filename", Input: "notes.md", Expected: true},
		{Name: "filename with dash", Input: "my-file.txt", Expected: true},
		{Name: "filename with underscore", Input: "my_file.txt", Expected: true},
		{Name: "uppercase", Input: "README.MD", Expected: true},
		{Name: "numbers", Input: "file123.txt", Expected: true},
		{Name: "empty", Input: "", Expected: false},
		{Name: "parent traversal", Input: "../etc/passwd", Expected: false},
		{Name: "hidden traversal", Input: "..secret", Expected: false},
		{Name: "forward slash", Input: "foo/bar.md", Expected: false},
		{Name: "backslash", Input: "foo\\bar.md", Expected: false},
		{Name: "hidden file", Input: ".gitignore", Expected: false},
		{Name: "just dots", Input: "..", Expected: false},
		{Name: "current dir", Input: ".", Expected: false},

		{Name: "fullwidth slash", Input: "foo／bar.md", Expected: false},
		{Name: "fullwidth backslash", Input: "foo＼bar.md", Expected: false},
		{Name: "fullwidth period traversal", Input: "．．／etc", Expected: false},
		{Name: "division slash", Input: "foo∕bar.md", Expected: false},
		{Name: "fraction slash", Input: "foo⁄bar.md", Expected: false},
		{Name: "null byte", Input: "file\x00.txt", Expected: false},
		{Name: "unicode dot", Input: "．secret", Expected: false},
		{Name: "cyrillic lookalike", Input: "tеst.md", Expected: false},
		{Name: "zero width char", Input: "test\u200b.md", Expected: false},
		{Name: "rtl override", Input: "test\u202e.md", Expected: false},
		{Name: "space", Input: "my file.txt", Expected: false},
		{Name: "newline", Input: "file\n.txt", Expected: false},
		{Name: "tab", Input: "file\t.txt", Expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			got := ValidFilePath(tt.Input)
			if got != tt.Expected {
				t.Errorf("ValidFilePath(%q) = %v, want %v", tt.Input, got, tt.Expected)
			}
		})
	}
}
