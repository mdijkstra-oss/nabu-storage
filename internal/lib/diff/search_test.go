package diff

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	th "hermes-relay/internal/lib/test-helpers"
)

type expectedMatch struct {
	Start   int    `json:"start"`
	End     int    `json:"end"`
	Fuzzy   bool   `json:"fuzzy"`
	Content string `json:"content,omitempty"`
}

type scenario struct {
	Name     string `json:"name"`
	Needle   string `json:"needle"`
	Expected struct {
		Matches []expectedMatch `json:"matches"`
	} `json:"expected"`
}

func scenariosDir() string {
	return filepath.Join("..", "..", "..", "..", "nabu-theatron", "app", "lib", "diff", "scenarios")
}

func loadContent(t *testing.T) string {
	data, err := os.ReadFile(filepath.Join(scenariosDir(), "content.md"))
	if err != nil {
		t.Fatalf("failed to load content.md: %v", err)
	}
	return string(data)
}

func loadScenarios(t *testing.T) []scenario {
	dir := scenariosDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read scenarios dir: %v", err)
	}

	var scenarios []scenario
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("failed to read %s: %v", entry.Name(), err)
		}
		var s scenario
		if err := json.Unmarshal(data, &s); err != nil {
			t.Fatalf("failed to parse %s: %v", entry.Name(), err)
		}
		scenarios = append(scenarios, s)
	}
	return scenarios
}

func TestFindMatches(t *testing.T) {
	content := loadContent(t)
	scenarios := loadScenarios(t)

	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {
			matches := FindMatches(content, s.Needle)

			th.AssertEqual(t, len(matches), len(s.Expected.Matches), "match count")

			for i, match := range matches {
				exp := s.Expected.Matches[i]
				th.AssertEqual(t, match.Start, exp.Start, "start")
				th.AssertEqual(t, match.End, exp.End, "end")
				th.AssertEqual(t, match.Fuzzy, exp.Fuzzy, "fuzzy")

				if exp.Content != "" {
					th.AssertEqual(t, GetMatchedText(content, match), exp.Content, "content")
				}
			}
		})
	}
}
