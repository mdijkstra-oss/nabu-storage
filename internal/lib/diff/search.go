package diff

import "strings"

type Match struct {
	Start int
	End   int
	Fuzzy bool
}

func FindMatches(content, needle string) []Match {
	contentLines := toLines(content)
	needleLines := toLines(needle)

	if len(needleLines) == 0 {
		return nil
	}
	if len(needleLines) > len(contentLines) {
		return nil
	}

	exactMatches := findExactMatches(contentLines, needleLines)
	if len(exactMatches) > 0 {
		return exactMatches
	}

	return findFuzzyMatches(contentLines, needleLines)
}

func GetMatchedText(content string, match Match) string {
	lines := toLines(content)
	return strings.Join(lines[match.Start:match.End+1], "\n")
}

const similarityThreshold = 0.9

func toLines(text string) []string {
	return strings.Split(text, "\n")
}

func findExactMatches(contentLines, needleLines []string) []Match {
	var matches []Match
	needleText := strings.Join(needleLines, "\n")

	for i := 0; i <= len(contentLines)-len(needleLines); i++ {
		slice := contentLines[i : i+len(needleLines)]
		if strings.Join(slice, "\n") == needleText {
			matches = append(matches, Match{Start: i, End: i + len(needleLines) - 1, Fuzzy: false})
		}
	}

	return matches
}

type scoredMatch struct {
	match Match
	score float64
}

func findFuzzyMatches(contentLines, needleLines []string) []Match {
	var scored []scoredMatch

	for i := 0; i <= len(contentLines)-len(needleLines); i++ {
		score := blockSimilarity(contentLines, needleLines, i)
		if score >= similarityThreshold {
			scored = append(scored, scoredMatch{
				match: Match{Start: i, End: i + len(needleLines) - 1, Fuzzy: true},
				score: score,
			})
		}
	}

	sortByScoreDesc(scored)

	matches := make([]Match, len(scored))
	for i, s := range scored {
		matches[i] = s.match
	}
	return matches
}

func sortByScoreDesc(matches []scoredMatch) {
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].score > matches[i].score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}

func blockSimilarity(contentLines, needleLines []string, startIndex int) float64 {
	var totalScore float64

	for i := 0; i < len(needleLines); i++ {
		score := lineSimilarity(contentLines[startIndex+i], needleLines[i])
		if score < similarityThreshold {
			return 0
		}
		totalScore += score
	}

	return totalScore / float64(len(needleLines))
}

func lineSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}

	// Normalize whitespace for comparison (trim leading/trailing and collapse internal)
	aNorm := normalizeWhitespace(a)
	bNorm := normalizeWhitespace(b)

	if aNorm == bNorm {
		return 1.0
	}
	if len(aNorm) == 0 && len(bNorm) == 0 {
		return 1.0
	}

	dist := levenshteinDistance(aNorm, bNorm)
	maxLen := len(aNorm)
	if len(bNorm) > maxLen {
		maxLen = len(bNorm)
	}

	return 1.0 - float64(dist)/float64(maxLen)
}

func normalizeWhitespace(s string) string {
	return strings.TrimSpace(s)
}

func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	aRunes := []rune(a)
	bRunes := []rune(b)

	prev := make([]int, len(bRunes)+1)
	curr := make([]int, len(bRunes)+1)

	for j := 0; j <= len(bRunes); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(aRunes); i++ {
		curr[0] = i
		for j := 1; j <= len(bRunes); j++ {
			cost := 0
			if aRunes[i-1] != bRunes[j-1] {
				cost = 1
			}
			curr[j] = min(
				prev[j]+1,
				curr[j-1]+1,
				prev[j-1]+cost,
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(bRunes)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
