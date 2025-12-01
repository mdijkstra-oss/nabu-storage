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
			{ID: "section-1", CodeID: "code-1", Text: "Climate change"},
			{ID: "section-2", CodeID: "code-2", Text: "impacts"},
			{ID: "section-3", CodeID: "code-3", Text: "global warming"},
		},
	}

	tests := []struct {
		Name     string
		Input    []string
		Expected float64
	}{
		{Name: "All codes coverage", Input: nil, Expected: 0.92},
		{Name: "Single code id", Input: []string{"code-1"}, Expected: 0.37},
		{Name: "Multiple code ids", Input: []string{"code-1", "code-3"}, Expected: 0.74},
		{Name: "Code id not in chunk returns 0.0", Input: []string{"code-99"}, Expected: 0.0},
	}

	test_helpers.RunFunctionTests[[]string, float64, float64](t, tests, func(codeIDs []string) float64 {
		return CalculateChunkCoverage(chunk, codeIDs)
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
	codeIDs     []string
}

func TestFilterChunksByCoverage(t *testing.T) {
	chunks := []file.Chunk{
		{ID: "1", Content: "AAAAAAAAAA", Codes: []file.CodedSection{{ID: "s1", CodeID: "code-1", Text: "AA"}}},
		{ID: "2", Content: "BBBBBBBBBB", Codes: []file.CodedSection{{ID: "s2", CodeID: "code-2", Text: "BBBBB"}}},
		{ID: "3", Content: "CCCCCCCCCC", Codes: []file.CodedSection{{ID: "s3", CodeID: "code-1", Text: "CCCCCCCC"}}},
		{ID: "4", Content: "DDDDDDDDDD", Codes: []file.CodedSection{
			{ID: "s4", CodeID: "code-1", Text: "DDDDD"},
			{ID: "s5", CodeID: "code-2", Text: "DDDDD"},
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
		{Name: "With specific code ids", Input: coverageFilterInput{floatPtr(0.5), nil, []string{"code-1"}}, Expected: []string{"3", "4"}},
		{Name: "Nil thresholds returns all chunks", Input: coverageFilterInput{nil, nil, nil}, Expected: []string{"1", "2", "3", "4", "5"}},
		{Name: "Filter by code ids with no coverage threshold", Input: coverageFilterInput{nil, nil, []string{"code-2"}}, Expected: []string{"1", "2", "3", "4", "5"}},
		{Name: "High min threshold filters most chunks", Input: coverageFilterInput{floatPtr(0.8), nil, nil}, Expected: []string{"3", "4"}},
		{Name: "Low max threshold filters most chunks", Input: coverageFilterInput{nil, floatPtr(0.3), nil}, Expected: []string{"1", "5"}},
		{Name: "Multiple code ids filter", Input: coverageFilterInput{floatPtr(0.4), nil, []string{"code-1", "code-2"}}, Expected: []string{"2", "3", "4"}},
	}

	test_helpers.RunFunctionTests(t, tests, func(input coverageFilterInput) []file.Chunk {
		return FilterChunksByCoverage(chunks, input.minCoverage, input.maxCoverage, input.codeIDs)
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
