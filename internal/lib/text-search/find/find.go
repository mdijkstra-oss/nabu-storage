package find

import "strings"

// MIN_OVERLAP at 1.0 requires exact token matching while still handling whitespace/punctuation
// variations via tokenization. Lower values (e.g. 0.8) cause boundary precision issues where
// non-matching tokens at the start of windows get included in returned indices.
// Punctuation at word boundaries is preserved in the returned range.
const MIN_OVERLAP = 1.0

// FindRange assumes small chunks of text, (like half to full page perhaps) multi-word
// It's fast, but looses accuracy on larger texts and smaller needles
func FindRange(needle, chunk string) (start, end int, found bool) {
	needleTokens := tokenize(needle)
	if len(needleTokens) == 0 {
		return 0, 0, false
	}

	// For fuzzy matching, we need variable-length windows
	// We slide through the text and expand windows to match token count
	for i := 0; i < len(chunk); i++ {
		// Skip leading whitespace - start window at actual word boundary
		if isWhitespace(chunk[i]) {
			continue
		}

		// Try to find a window starting at position i that has the right number of tokens
		windowEnd := findWindowEnd(chunk, i, len(needleTokens))
		if windowEnd == -1 {
			// Can't find enough tokens from this position
			continue
		}

		window := chunk[i:windowEnd]
		windowTokens := tokenize(window)

		// Check if we have the right number of tokens and sufficient overlap
		if len(windowTokens) == len(needleTokens) && tokenOverlap(needleTokens, windowTokens) >= MIN_OVERLAP {
			return i, windowEnd, true
		}
	}
	return 0, 0, false
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// findWindowEnd finds the end position of a window that contains exactly tokenCount tokens
// Returns -1 if not enough tokens are available
func findWindowEnd(text string, start int, tokenCount int) int {
	if start >= len(text) {
		return -1
	}

	tokensFound := 0
	inWord := false

	for i := start; i < len(text); i++ {
		char := text[i]
		isWS := isWhitespace(char)

		if !isWS && !inWord {
			// Start of a new word
			inWord = true
			tokensFound++
			if tokensFound == tokenCount {
				// Find the end of this word
				for j := i + 1; j <= len(text); j++ {
					if j == len(text) {
						return j
					}
					if isWhitespace(text[j]) {
						return j
					}
				}
				return len(text)
			}
		} else if isWS && inWord {
			// End of a word
			inWord = false
		}
	}

	// If we found exactly the right number of tokens and we're at the end
	if tokensFound == tokenCount {
		return len(text)
	}

	return -1
}

func tokenize(s string) []string {
	words := strings.Fields(strings.ToLower(s))
	// Strip punctuation from each token
	result := make([]string, 0, len(words))
	for _, word := range words {
		cleaned := stripPunctuation(word)
		if cleaned != "" {
			result = append(result, cleaned)
		}
	}
	return result
}

func stripPunctuation(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Keep alphanumeric characters
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result.WriteByte(c)
		}
	}
	return result.String()
}

func tokenOverlap(needle, window []string) float64 {
	if len(needle) == 0 {
		return 0
	}

	// Create frequency maps for both needle and window
	needleFreq := make(map[string]int)
	for _, t := range needle {
		needleFreq[t]++
	}

	windowFreq := make(map[string]int)
	for _, t := range window {
		windowFreq[t]++
	}

	// Count how many needle tokens are satisfied by window tokens
	satisfied := 0
	for token, needCount := range needleFreq {
		windowCount := windowFreq[token]
		if windowCount > 0 {
			// Take the minimum - we can only satisfy as many as we have
			if windowCount >= needCount {
				satisfied += needCount
			} else {
				satisfied += windowCount
			}
		}
	}

	return float64(satisfied) / float64(len(needle))
}
