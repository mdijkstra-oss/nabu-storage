package find

import (
	"encoding/json"
	"flag"
	"os"
	"testing"
)

var updateBaseline = flag.Bool("update-baseline", false, "Update baseline.json with current results")

type ParsedRequest struct {
	ChunkContent string   `json:"chunk_content"`
	FoundTexts   []string `json:"found_texts"`
	FailedTexts  []string `json:"failed_texts"`
}

type Baseline struct {
	PassingTexts map[string]int `json:"passing_texts"`
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

func loadBaseline(path string) (map[string]int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]int), nil
		}
		return nil, err
	}
	var baseline Baseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return nil, err
	}
	return baseline.PassingTexts, nil
}

func saveBaseline(path string, passing map[string]int) error {
	baseline := Baseline{PassingTexts: passing}
	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
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

	baselineMap, err := loadBaseline("baseline.json")
	if err != nil {
		t.Fatalf("Failed to load baseline.json: %v", err)
	}

	currentPassing := make(map[string]int)
	var regressions, improvements int

	for chunkIdx, req := range requests {
		for _, needle := range req.FoundTexts {
			_, found := Find(needle, req.ChunkContent)
			baselineChunk, wasInBaseline := baselineMap[needle]

			if found {
				currentPassing[needle] = chunkIdx
				if !wasInBaseline {
					improvements++
				}
			} else if wasInBaseline && baselineChunk == chunkIdx {
				regressions++
				t.Errorf("REGRESSION: baseline text no longer found: %q", truncate(needle, 80))
			}
		}

		for _, needle := range req.FailedTexts {
			_, found := Find(needle, req.ChunkContent)
			baselineChunk, wasInBaseline := baselineMap[needle]

			if found {
				currentPassing[needle] = chunkIdx
				if !wasInBaseline {
					improvements++
				}
			} else if wasInBaseline && baselineChunk == chunkIdx {
				regressions++
				t.Errorf("REGRESSION: baseline text no longer found: %q", truncate(needle, 80))
			}
		}
	}

	t.Logf("\n=== REGRESSION SUMMARY ===")
	t.Logf("Baseline texts:  %d", len(baselineMap))
	t.Logf("Now passing:     %d", len(currentPassing))
	t.Logf("Regressions:     %d", regressions)
	t.Logf("Improvements:    %d", improvements)
	t.Logf("==========================")

	if *updateBaseline {
		if err := saveBaseline("baseline.json", currentPassing); err != nil {
			t.Fatalf("Failed to save baseline: %v", err)
		}
		t.Logf("Baseline updated with %d passing texts", len(currentPassing))
	}
}
