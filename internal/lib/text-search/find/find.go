package find

import (
	"strings"
	"unicode"
)

const MIN_OVERLAP = 1.0

type wordPosition struct {
	start, end int
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
	needleTokens := tokenize(needle)
	if len(needleTokens) == 0 {
		return "", false
	}

	words := extractWords(chunk)
	if len(words) < len(needleTokens) {
		return "", false
	}

	for i := 0; i <= len(words)-len(needleTokens); i++ {
		window := words[i : i+len(needleTokens)]
		start := window[0].start
		end := window[len(window)-1].end

		windowText := chunk[start:end]
		windowTokens := tokenize(windowText)

		if len(windowTokens) == len(needleTokens) && tokenOverlap(needleTokens, windowTokens) >= MIN_OVERLAP && isSubsequence(needleTokens, windowTokens) {
			start, end = expandToWordBoundaries(chunk, start, end)
			start, end = BalanceMarkdownTags(chunk, start, end)
			return chunk[start:end], true
		}
	}

	return "", false
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
