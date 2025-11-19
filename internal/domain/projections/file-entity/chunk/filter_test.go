package chunk

import (
	"hermes-relay/internal/domain/entities/file"
	test_helpers "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func TestCalculateChunkCoverage(t *testing.T) {
	chunk := file.Chunk{
		ID:      "1",
		Content: "Climate change impacts global warming.",
		Codes: []file.CodedSection{
			{CodeSlug: "topic:climate", CodeID: "code-1", CodedSectionAttributes: file.CodedSectionAttributes{Text: "Climate change"}},
			{CodeSlug: "emotion:concern", CodeID: "code-2", CodedSectionAttributes: file.CodedSectionAttributes{Text: "impacts"}},
			{CodeSlug: "topic:temperature", CodeID: "code-3", CodedSectionAttributes: file.CodedSectionAttributes{Text: "global warming"}},
		},
	}

	tests := []struct {
		Name     string
		Input    []string
		Expected float64
	}{
		{Name: "All codes coverage", Input: nil, Expected: 0.92},
		{Name: "Single code slug", Input: []string{"topic:climate"}, Expected: 0.37},
		{Name: "Multiple code slugs", Input: []string{"topic:climate", "topic:temperature"}, Expected: 0.74},
		{Name: "Code slug not in chunk returns 0.0", Input: []string{"topic:economy"}, Expected: 0.0},
	}

	test_helpers.RunFunctionTests[[]string, float64, float64](t, tests, func(codeSlugs []string) float64 {
		return CalculateChunkCoverage(chunk, codeSlugs)
	})
}

func TestFilterChunksByText(t *testing.T) {
	chunks := []file.Chunk{
		{ID: "1", Content: "Climate change impacts global warming.", Codes: []file.CodedSection{}},
		{ID: "2", Content: "Economic policies affect market stability.", Codes: []file.CodedSection{}},
		{ID: "3", Content: "Healthcare systems need reform urgently.", Codes: []file.CodedSection{}},
	}

	tests := []struct {
		Name     string
		Input    string
		Expected []string
	}{
		{Name: "Exact match finds chunk", Input: "Climate change", Expected: []string{"1"}},
		{Name: "Fuzzy match finds chunk", Input: "climate", Expected: []string{"1"}},
		{Name: "Multi-word match", Input: "market stability", Expected: []string{"2"}},
		{Name: "No match returns empty", Input: "quantum physics", Expected: []string{}},
		{Name: "Empty search text returns all chunks", Input: "", Expected: []string{"1", "2", "3"}},
		{Name: "Case insensitive match", Input: "HEALTHCARE", Expected: []string{"3"}},
		{Name: "Single word match", Input: "economic", Expected: []string{"2"}},
	}

	test_helpers.RunFunctionTests(t, tests, func(searchText string) []file.Chunk {
		return FilterChunksByText(chunks, searchText)
	}, extractIDs)
}

type coverageFilterInput struct {
	minCoverage *float64
	maxCoverage *float64
	codeSlugs   []string
}

func TestFilterChunksByCoverage(t *testing.T) {
	chunks := []file.Chunk{
		{ID: "1", Content: "AAAAAAAAAA", Codes: []file.CodedSection{{CodeSlug: "topic:climate", CodeID: "code-1", CodedSectionAttributes: file.CodedSectionAttributes{Text: "AA"}}}},
		{ID: "2", Content: "BBBBBBBBBB", Codes: []file.CodedSection{{CodeSlug: "topic:economy", CodeID: "code-2", CodedSectionAttributes: file.CodedSectionAttributes{Text: "BBBBB"}}}},
		{ID: "3", Content: "CCCCCCCCCC", Codes: []file.CodedSection{{CodeSlug: "topic:climate", CodeID: "code-3", CodedSectionAttributes: file.CodedSectionAttributes{Text: "CCCCCCCC"}}}},
		{ID: "4", Content: "DDDDDDDDDD", Codes: []file.CodedSection{
			{CodeSlug: "topic:climate", CodeID: "code-4", CodedSectionAttributes: file.CodedSectionAttributes{Text: "DDDDD"}},
			{CodeSlug: "topic:economy", CodeID: "code-5", CodedSectionAttributes: file.CodedSectionAttributes{Text: "DDDDD"}},
		}},
		{ID: "5", Content: "EEEEEEEEEE", Codes: []file.CodedSection{}},
	}

	tests := []struct {
		Name     string
		Input    coverageFilterInput
		Expected []string
	}{
		{Name: "Min coverage only", Input: coverageFilterInput{floatPtr(0.5), nil, nil}, Expected: []string{"2", "3", "4"}},
		{Name: "Max coverage only", Input: coverageFilterInput{nil, floatPtr(0.5), nil}, Expected: []string{"1", "2", "5"}},
		{Name: "Both min and max coverage", Input: coverageFilterInput{floatPtr(0.3), floatPtr(0.7), nil}, Expected: []string{"2"}},
		{Name: "With specific code slugs", Input: coverageFilterInput{floatPtr(0.5), nil, []string{"topic:climate"}}, Expected: []string{"3", "4"}},
		{Name: "Nil thresholds returns all chunks", Input: coverageFilterInput{nil, nil, nil}, Expected: []string{"1", "2", "3", "4", "5"}},
		{Name: "Filter by code slugs with no coverage threshold", Input: coverageFilterInput{nil, nil, []string{"topic:economy"}}, Expected: []string{"1", "2", "3", "4", "5"}},
		{Name: "High min threshold filters most chunks", Input: coverageFilterInput{floatPtr(0.8), nil, nil}, Expected: []string{"3", "4"}},
		{Name: "Low max threshold filters most chunks", Input: coverageFilterInput{nil, floatPtr(0.3), nil}, Expected: []string{"1", "5"}},
		{Name: "Multiple code slugs filter", Input: coverageFilterInput{floatPtr(0.4), nil, []string{"topic:climate", "topic:economy"}}, Expected: []string{"2", "3", "4"}},
	}

	test_helpers.RunFunctionTests(t, tests, func(input coverageFilterInput) []file.Chunk {
		return FilterChunksByCoverage(chunks, input.minCoverage, input.maxCoverage, input.codeSlugs)
	}, extractIDs)
}

func floatPtr(f float64) *float64 {
	return &f
}

func extractIDs(chunks []file.Chunk) []string {
	ids := make([]string, len(chunks))
	for i, chunk := range chunks {
		ids[i] = chunk.ID
	}
	return ids
}
