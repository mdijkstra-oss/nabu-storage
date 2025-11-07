package textsearch

import (
	"hermes-relay/internal/lib/text-search/find"
	"math"
)

func CalculateCoverage(text string, subTexts []string) float64 {
	if len(text) == 0 {
		return 0.0
	}

	totalChars := 0
	for _, subText := range subTexts {
		totalChars += len(subText)
	}

	coverage := float64(totalChars) / float64(len(text))
	return math.Round(coverage*100) / 100
}

func ContainsText(text, searchText string) bool {
	_, _, found := find.FindRange(searchText, text)
	return found
}
