package find

import "strings"

var markdownDelimiters = []string{"**", "__", "*", "_", "`"}

func BalanceMarkdownTags(text string, start, end int) (int, int) {
	for _, delim := range markdownDelimiters {
		start, end = balanceDelimiter(text, start, end, delim)
	}
	return start, end
}

func balanceDelimiter(text string, start, end int, delim string) (int, int) {
	selection := text[start:end]
	count := strings.Count(selection, delim)

	if count%2 == 0 {
		return start, end
	}

	openingBefore := findDelimiterBefore(text, start, delim)
	closingAfter := findDelimiterAfter(text, end, delim)

	if openingBefore >= 0 && closingAfter >= 0 {
		distToBefore := start - openingBefore
		distToAfter := closingAfter - end
		if distToBefore <= distToAfter {
			return openingBefore, end
		}
		return start, closingAfter + len(delim)
	}

	if openingBefore >= 0 {
		return openingBefore, end
	}

	if closingAfter >= 0 {
		return start, closingAfter + len(delim)
	}

	return start, end
}

func findDelimiterBefore(text string, pos int, delim string) int {
	searchArea := text[:pos]
	idx := strings.LastIndex(searchArea, delim)
	return idx
}

func findDelimiterAfter(text string, pos int, delim string) int {
	if pos >= len(text) {
		return -1
	}
	searchArea := text[pos:]
	idx := strings.Index(searchArea, delim)
	if idx < 0 {
		return -1
	}
	return pos + idx
}
