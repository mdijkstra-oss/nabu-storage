package textsearch

import "strings"

const MIN_OVERLAP = 0.8

// FindRange assumes small chunks of text, (like half to full page perhaps) multi-word
// It's fast, but looses accuracy on larger texts and smaller needles
func FindRange(needle, chunk string) (start, end int, found bool) {
	needleTokens := tokenize(needle)

	// Slide window over original text - small text so good enough
	for i := 0; i < len(chunk)-len(needle); i++ {
		window := chunk[i : i+len(needle)]
		if tokenOverlap(needleTokens, tokenize(window)) > MIN_OVERLAP {
			return i, i + len(needle), true
		}
	}
	return 0, 0, false
}

func tokenize(s string) []string {
	return strings.Fields(strings.ToLower(s))
}

func tokenOverlap(a, b []string) float64 {
	aSet := make(map[string]bool)
	for _, t := range a {
		aSet[t] = true
	}

	matches := 0
	for _, t := range b {
		if aSet[t] {
			matches++
		}
	}
	return float64(matches) / float64(len(a))
}
