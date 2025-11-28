package find

import (
	"encoding/json"
	"os"
	"testing"
)

type ParsedRequest struct {
	ChunkContent string   `json:"chunk_content"`
	FoundTexts   []string `json:"found_texts"`
	FailedTexts  []string `json:"failed_texts"`
}

type RegressionStats struct {
	FoundTextsTotal  int
	FoundTextsFound  int
	FailedTextsTotal int
	FailedTextsFound int
}

func loadParsedRequests(path string) ([]ParsedRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var requests []ParsedRequest
	err = json.Unmarshal(data, &requests)
	return requests, err
}

func toSet(items []string) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for _, item := range items {
		set[item] = struct{}{}
	}
	return set
}

func runRegressionOnRequest(t *testing.T, req ParsedRequest) RegressionStats {
	var stats RegressionStats
	failedSet := toSet(req.FailedTexts)

	for _, needle := range req.FoundTexts {
		_, inFailedSet := failedSet[needle]
		if inFailedSet {
			continue
		}
		stats.FoundTextsTotal++
		_, found := Find(needle, req.ChunkContent)
		if found {
			stats.FoundTextsFound++
		} else {
			t.Errorf("REGRESSION: previously found text not found: %q", truncate(needle, 80))
		}
	}

	for _, needle := range req.FailedTexts {
		stats.FailedTextsTotal++
		_, found := Find(needle, req.ChunkContent)
		if found {
			stats.FailedTextsFound++
		}
	}

	return stats
}

func aggregateStats(a, b RegressionStats) RegressionStats {
	return RegressionStats{
		FoundTextsTotal:  a.FoundTextsTotal + b.FoundTextsTotal,
		FoundTextsFound:  a.FoundTextsFound + b.FoundTextsFound,
		FailedTextsTotal: a.FailedTextsTotal + b.FailedTextsTotal,
		FailedTextsFound: a.FailedTextsFound + b.FailedTextsFound,
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func TestFind_Regression(t *testing.T) {
	requests, err := loadParsedRequests("parsed-requests.json")
	if err != nil {
		t.Fatalf("Failed to load parsed-requests.json: %v", err)
	}

	var total RegressionStats
	for i, req := range requests {
		stats := runRegressionOnRequest(t, req)
		total = aggregateStats(total, stats)
		t.Logf("Chunk %d: found_texts %d/%d, failed_texts now found %d/%d",
			i, stats.FoundTextsFound, stats.FoundTextsTotal,
			stats.FailedTextsFound, stats.FailedTextsTotal)
	}

	t.Logf("\n=== REGRESSION SUMMARY ===")
	t.Logf("Previously found (must stay found): %d/%d (%.1f%%)",
		total.FoundTextsFound, total.FoundTextsTotal,
		percentage(total.FoundTextsFound, total.FoundTextsTotal))
	t.Logf("Previously failed (now found):      %d/%d (%.1f%%)",
		total.FailedTextsFound, total.FailedTextsTotal,
		percentage(total.FailedTextsFound, total.FailedTextsTotal))
	t.Logf("==========================")
}

func percentage(found, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(found) / float64(total) * 100
}
