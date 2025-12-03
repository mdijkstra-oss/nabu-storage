package find

import (
	"strings"
	"unicode"
)

const MIN_OVERLAP = 1.0

type wordPosition struct {
	start, end int
}

func headingLevel(s string) int {
	trimmed := strings.TrimLeft(s, " \t")
	count := 0
	for _, r := range trimmed {
		if r == '#' {
			count++
		} else {
			break
		}
	}
	if count > 0 && count < len(trimmed) && (trimmed[count] == ' ' || trimmed[count] == '\t') {
		return count
	}
	return 0
}

func matchIsAtHeading(chunk string, matchStart, requiredLevel int) bool {
	lineStart := matchStart
	for lineStart > 0 && chunk[lineStart-1] != '\n' {
		lineStart--
	}
	lineContent := chunk[lineStart:]
	return headingLevel(lineContent) == requiredLevel
}

func extractWords(text string) []wordPosition {
	var words []wordPosition
	inWord := false
	var start int

	for i, r := range text {
		isWordChar := unicode.IsLetter(r) || unicode.IsNumber(r) || r == '\'' || r == '\u2019'

		if isWordChar && !inWord {
			start = i
			inWord = true
		} else if !isWordChar && inWord {
			words = append(words, wordPosition{start, i})
			inWord = false
		}
	}

	if inWord {
		words = append(words, wordPosition{start, len(text)})
	}

	return words
}

func expandToWordBoundaries(text string, start, end int) (int, int) {
	for start > 0 && !isSpaceOrNBSP(rune(text[start-1])) {
		start--
	}
	for end < len(text) && !isSpaceOrNBSP(rune(text[end])) {
		end++
	}
	return start, end
}

func isSpaceOrNBSP(r rune) bool {
	return unicode.IsSpace(r) || r == '\u00A0'
}

func Find(needle, chunk string) (text string, found bool) {
	matches := FindN(needle, chunk, 1)
	if len(matches) == 0 {
		return "", false
	}
	return matches[0].Text, true
}

func FindAll(needle, chunk string) []Match {
	return FindN(needle, chunk, 0)
}

func tokenize(s string) []string {
	s = NormalizeText(s)

	var tokens []string
	var currentWord strings.Builder

	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			currentWord.WriteRune(r)
		} else {
			if currentWord.Len() > 0 {
				tokens = append(tokens, currentWord.String())
				currentWord.Reset()
			}
		}
	}

	if currentWord.Len() > 0 {
		tokens = append(tokens, currentWord.String())
	}

	return tokens
}

func tokenOverlap(needle, window []string) float64 {
	if len(needle) == 0 {
		return 0
	}

	needleFreq := make(map[string]int)
	for _, t := range needle {
		needleFreq[t]++
	}

	windowFreq := make(map[string]int)
	for _, t := range window {
		windowFreq[t]++
	}

	satisfied := 0
	for token, needCount := range needleFreq {
		windowCount := windowFreq[token]
		if windowCount > 0 {
			if windowCount >= needCount {
				satisfied += needCount
			} else {
				satisfied += windowCount
			}
		}
	}

	return float64(satisfied) / float64(len(needle))
}

func isSubsequence(needle, window []string) bool {
	needleIdx := 0
	for _, token := range window {
		if needleIdx < len(needle) && token == needle[needleIdx] {
			needleIdx++
		}
	}
	return needleIdx == len(needle)
}

type Match struct {
	Text  string
	Start int
	End   int
}

func expandToLineStart(text string, pos int) int {
	for pos > 0 && text[pos-1] != '\n' {
		pos--
	}
	return pos
}

func FindN(needle, chunk string, limit int) []Match {
	needleTokens := tokenize(needle)
	if len(needleTokens) == 0 {
		return nil
	}

	requiredHeadingLevel := headingLevel(needle)

	words := extractWords(chunk)
	if len(words) < len(needleTokens) {
		return nil
	}

	var matches []Match
	lastEnd := -1

	for i := 0; i <= len(words)-len(needleTokens); i++ {
		window := words[i : i+len(needleTokens)]
		start := window[0].start
		end := window[len(window)-1].end

		if start < lastEnd {
			continue
		}

		windowText := chunk[start:end]
		windowTokens := tokenize(windowText)

		if len(windowTokens) == len(needleTokens) && tokenOverlap(needleTokens, windowTokens) >= MIN_OVERLAP && isSubsequence(needleTokens, windowTokens) {
			if requiredHeadingLevel > 0 && !matchIsAtHeading(chunk, start, requiredHeadingLevel) {
				continue
			}
			if requiredHeadingLevel > 0 {
				start = expandToLineStart(chunk, start)
			}
			start, end = expandToWordBoundaries(chunk, start, end)
			start, end = BalanceMarkdownTags(chunk, start, end)
			matches = append(matches, Match{Text: chunk[start:end], Start: start, End: end})
			if limit > 0 && len(matches) >= limit {
				return matches
			}
			lastEnd = end
		}
	}

	return matches
}

func ExtractContext(text string, matchStart, matchEnd, sentenceCount int) string {
	isSentenceEnd := func(b byte) bool {
		return b == '.' || b == '!' || b == '?' || b == '\n'
	}

	contextStart := matchStart
	for i := 0; i < sentenceCount; i++ {
		for contextStart > 0 && !isSentenceEnd(text[contextStart-1]) {
			contextStart--
		}
		if contextStart > 0 {
			contextStart--
		}
	}
	for contextStart > 0 && !isSentenceEnd(text[contextStart-1]) {
		contextStart--
	}

	contextEnd := matchEnd
	for i := 0; i < sentenceCount; i++ {
		for contextEnd < len(text) && isSentenceEnd(text[contextEnd]) {
			contextEnd++
		}
		for contextEnd < len(text) && !isSentenceEnd(text[contextEnd]) {
			contextEnd++
		}
		if contextEnd < len(text) {
			contextEnd++
		}
	}

	return strings.TrimSpace(text[contextStart:contextEnd])
}

func Replace(content, oldText, newText string) string {
	if oldText == "" {
		return content
	}
	return strings.Replace(content, oldText, newText, 1)
}

func CountWords(text string) int {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return 0
	}

	wordCount := 0
	inWord := false

	for _, r := range trimmed {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			if !inWord {
				wordCount++
				inWord = true
			}
		} else {
			inWord = false
		}
	}

	return wordCount
}
