package find

import "testing"

func TestHeadingLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"h1", "# Heading", 1},
		{"h2", "## Heading", 2},
		{"h3", "### Heading", 3},
		{"h6", "###### Heading", 6},
		{"h2 with tab", "##\tHeading", 2},
		{"leading space", "  ## Heading", 2},
		{"no space after hash", "##Heading", 0},
		{"plain text", "Heading", 0},
		{"hash in middle", "This is # not a heading", 0},
		{"empty", "", 0},
		{"only hashes", "###", 0},
		{"hash with trailing space only", "## ", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := headingLevel(tt.input)
			if got != tt.want {
				t.Errorf("headingLevel(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestMatchIsAtHeading(t *testing.T) {
	tests := []struct {
		name          string
		chunk         string
		matchStart    int
		requiredLevel int
		want          bool
	}{
		{"h2 at start", "## Heading\nContent", 0, 2, true},
		{"h2 after newline", "Content\n## Heading", 8, 2, true},
		{"wrong level", "## Heading", 0, 3, false},
		{"not a heading", "Heading", 0, 2, false},
		{"match mid-line", "Some ## text", 5, 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchIsAtHeading(tt.chunk, tt.matchStart, tt.requiredLevel)
			if got != tt.want {
				t.Errorf("matchIsAtHeading(%q, %d, %d) = %v, want %v",
					tt.chunk, tt.matchStart, tt.requiredLevel, got, tt.want)
			}
		})
	}
}

func TestFind_HeadingAware(t *testing.T) {
	tests := []FindTestCase{
		{
			name:     "heading needle matches heading",
			text:     "## My Heading\n\nSome content here",
			needle:   "## My Heading",
			wantFind: true,
		},
		{
			name:     "heading needle skips non-heading text",
			text:     "My Heading appears here\n\n## My Heading\n\nMore content",
			needle:   "## My Heading",
			wantFind: true,
		},
		{
			name:     "heading needle no match when no heading exists",
			text:     "My Heading appears here but not as a heading",
			needle:   "## My Heading",
			wantFind: false,
		},
		{
			name:     "h3 needle matches h3 not h2",
			text:     "## My Heading\n\n### My Heading\n\nContent",
			needle:   "### My Heading",
			wantFind: true,
		},
		{
			name:     "non-heading needle still works",
			text:     "## My Heading\n\nSome content here",
			needle:   "Some content here",
			wantFind: true,
		},
		{
			name:     "heading with extra content",
			text:     "## Introduction\n\nThis is the intro section",
			needle:   "## Introduction",
			wantFind: true,
		},
	}

	runFindTests(t, tests)
}
