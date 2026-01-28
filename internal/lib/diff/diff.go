package diff

import (
	"fmt"
	"strings"
)

type Hunk struct {
	OldText string
	NewText string
}

type Result struct {
	OK      bool
	Content string
	Error   string
}

func Ok(content string) Result {
	return Result{OK: true, Content: content}
}

func Fail(err string) Result {
	return Result{OK: false, Error: err}
}

func isHunkStart(line string) bool {
	return line == "@@" || strings.HasPrefix(line, "@@ ") ||
		line == "+@@" || strings.HasPrefix(line, "+@@ ")
}

func isAddLine(line string) bool {
	return strings.HasPrefix(line, "+")
}

func isRemoveLine(line string) bool {
	return strings.HasPrefix(line, "-")
}

func stripAddPrefix(line string) string {
	if strings.HasPrefix(line, "++") {
		return line[2:]
	}
	return line[1:]
}

type parseState struct {
	hunks      []Hunk
	currentOld strings.Builder
	currentNew strings.Builder
	inHunk     bool
	isAddFile  bool
}

func newParseState() *parseState {
	return &parseState{hunks: []Hunk{}}
}

func flushHunk(s *parseState) {
	if s.inHunk && (s.currentOld.Len() > 0 || s.currentNew.Len() > 0) {
		s.hunks = append(s.hunks, Hunk{
			OldText: s.currentOld.String(),
			NewText: s.currentNew.String(),
		})
	}
	s.currentOld.Reset()
	s.currentNew.Reset()
}

func isAddFileMarker(line string) bool {
	return strings.HasPrefix(line, "*** Add File:")
}

func isUpdateOrDeleteMarker(line string) bool {
	return strings.HasPrefix(line, "*** Update File:") || strings.HasPrefix(line, "*** Delete File:")
}

func isMetaLine(line string) bool {
	return strings.HasPrefix(line, "*** ")
}

func parseV4ADiff(patch string) []Hunk {
	lines := strings.Split(patch, "\n")
	state := newParseState()

	for _, line := range lines {
		parseLine(state, line)
	}

	flushHunk(state)
	return state.hunks
}

func parseLine(s *parseState, line string) {
	switch {
	case isAddFileMarker(line):
		flushHunk(s)
		s.isAddFile = true
		s.inHunk = true
	case isUpdateOrDeleteMarker(line):
		flushHunk(s)
		s.isAddFile = false
		s.inHunk = false
	case isHunkStart(line):
		flushHunk(s)
		s.inHunk = true
		s.isAddFile = false
	case isMetaLine(line):
		return
	default:
		parseContentLine(s, line)
	}
}

func parseContentLine(s *parseState, line string) {
	if !s.inHunk && (isAddLine(line) || isRemoveLine(line)) {
		s.inHunk = true
	}

	if !s.inHunk {
		return
	}

	switch {
	case isRemoveLine(line):
		s.currentOld.WriteString(line[1:])
		s.currentOld.WriteString("\n")
	case isAddLine(line):
		s.currentNew.WriteString(stripAddPrefix(line))
		s.currentNew.WriteString("\n")
	case !s.isAddFile:
		s.currentOld.WriteString(line)
		s.currentOld.WriteString("\n")
		s.currentNew.WriteString(line)
		s.currentNew.WriteString("\n")
	default:
		s.currentNew.WriteString(line)
		s.currentNew.WriteString("\n")
	}
}

const contextLines = 3

func countLines(content string) int {
	return len(strings.Split(content, "\n"))
}

func expandMatch(match Match, n, totalLines int) Match {
	start := match.Start - n
	if start < 0 {
		start = 0
	}
	end := match.End + n
	if end > totalLines-1 {
		end = totalLines - 1
	}
	return Match{Start: start, End: end, Fuzzy: match.Fuzzy}
}

func formatAmbiguousError(content string, matches []Match) string {
	total := countLines(content)
	var parts []string
	parts = append(parts, fmt.Sprintf("patch ambiguous: %d matches found", len(matches)))
	parts = append(parts, "")

	for i, m := range matches {
		exp := expandMatch(m, contextLines, total)
		text := GetMatchedText(content, exp)
		parts = append(parts, fmt.Sprintf("Match %d:", i+1))
		parts = append(parts, "───")
		parts = append(parts, text)
		parts = append(parts, "───")
	}

	parts = append(parts, "")
	parts = append(parts, "Include more surrounding lines to disambiguate. If none look right, verify you're targeting the correct location.")

	return strings.Join(parts, "\n")
}

func applyHunk(content string, hunk Hunk) Result {
	oldText := strings.TrimSuffix(hunk.OldText, "\n")
	newText := strings.TrimSuffix(hunk.NewText, "\n")

	if oldText == "" {
		needsSeparator := len(content) > 0 && !strings.HasSuffix(content, "\n")
		if needsSeparator {
			return Ok(content + "\n" + newText)
		}
		return Ok(content + newText)
	}

	matches := FindMatches(content, oldText)

	if len(matches) == 0 {
		preview := oldText
		if len(preview) > 50 {
			preview = preview[:50] + "..."
		}
		return Fail("patch context not found: \"" + preview + "\"")
	}

	if len(matches) > 1 {
		return Fail(formatAmbiguousError(content, matches))
	}

	matchedText := GetMatchedText(content, matches[0])
	return Ok(strings.Replace(content, matchedText, newText, 1))
}

func Apply(content, patch string) Result {
	hunks := parseV4ADiff(patch)

	result := content
	for _, hunk := range hunks {
		applied := applyHunk(result, hunk)
		if !applied.OK {
			return applied
		}
		result = applied.Content
	}

	return Ok(result)
}
