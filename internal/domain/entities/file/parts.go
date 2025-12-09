package file

import (
	"hermes-relay/internal/lib/text-search/find"
	"strings"
)

const DefaultPartSize = 12000

type FilePart struct {
	Content string         `json:"content"`
	Codes   []CodedSection `json:"codes"`
}

type partRange struct {
	start int
	end   int
	text  string
}

func SplitIntoParts(content string, codes []CodedSection, maxSize int) []FilePart {
	if content == "" {
		return []FilePart{}
	}

	ranges := splitAtNewlines(content, maxSize)
	return buildFileParts(ranges, codes, content)
}

func splitAtNewlines(content string, maxSize int) []partRange {
	if len(content) <= maxSize {
		return []partRange{{start: 0, end: len(content), text: content}}
	}

	ranges := []partRange{}
	start := 0

	for start < len(content) {
		end := start + maxSize
		if end >= len(content) {
			ranges = append(ranges, partRange{
				start: start,
				end:   len(content),
				text:  content[start:],
			})
			break
		}

		splitPoint := findNearestNewline(content, end)
		ranges = append(ranges, partRange{
			start: start,
			end:   splitPoint,
			text:  content[start:splitPoint],
		})
		start = splitPoint
	}

	return ranges
}

func findNearestNewline(content string, position int) int {
	if position >= len(content) {
		return len(content)
	}

	lookAhead := 500
	lookBehind := 500

	searchEnd := position + lookAhead
	if searchEnd > len(content) {
		searchEnd = len(content)
	}

	afterPos := strings.Index(content[position:searchEnd], "\n")
	if afterPos != -1 {
		return position + afterPos + 1
	}

	searchStart := position - lookBehind
	if searchStart < 0 {
		searchStart = 0
	}

	beforePos := strings.LastIndex(content[searchStart:position], "\n")
	if beforePos != -1 {
		return searchStart + beforePos + 1
	}

	return position
}

func buildFileParts(ranges []partRange, codes []CodedSection, fullContent string) []FilePart {
	codePositions := findCodePositions(codes, fullContent)

	parts := make([]FilePart, len(ranges))
	for i, r := range ranges {
		parts[i] = FilePart{
			Content: r.text,
			Codes:   codesInRange(codes, codePositions, r.start, r.end),
		}
	}
	return parts
}

func findCodePositions(codes []CodedSection, content string) map[string]int {
	positions := make(map[string]int, len(codes))

	for _, code := range codes {
		matches := find.FindAll(code.Text, content)
		if len(matches) > 0 {
			positions[code.ID] = matches[0].Start
		}
	}

	return positions
}

func codesInRange(codes []CodedSection, positions map[string]int, start, end int) []CodedSection {
	filtered := []CodedSection{}

	for _, code := range codes {
		pos, found := positions[code.ID]
		if found && pos >= start && pos < end {
			filtered = append(filtered, code)
		}
	}

	return filtered
}
