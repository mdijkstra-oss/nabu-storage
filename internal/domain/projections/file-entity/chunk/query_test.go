package chunk

import (
	"hermes-relay/internal/domain/entities/file"
	test_helpers "hermes-relay/internal/lib/test-helpers"
	"testing"
)

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func TestByChunk(t *testing.T) {
	file1 := file.BuildTestFile("file-1", file.FileData{
		ProjectID: "project-1",
		Name:      "test.txt",
	})
	file1.Chunks = []file.Chunk{
		{ID: "1", Content: "Climate change impacts", Codes: []file.CodedSection{
			{CodeSlug: "topic:climate", CodeID: "code-1", ID: "s1", Text: "Climate change"},
		}},
		{ID: "2", Content: "Economic policies affect", Codes: []file.CodedSection{
			{CodeSlug: "topic:economy", CodeID: "code-2", ID: "s1", Text: "Economic"},
		}},
		{ID: "3", Content: "Climate patterns shift", Codes: []file.CodedSection{
			{CodeSlug: "topic:climate", CodeID: "code-3", ID: "s1", Text: "Climate"},
		}},
	}

	tests := []struct {
		Name     string
		Input    ChunkQuery
		Expected *ChunkResult
	}{
		{
			Name:  "First chunk by default",
			Input: ChunkQuery{},
			Expected: &ChunkResult{
				Chunk:       file.Chunk{ID: "1", Content: "Climate change impacts", Codes: []file.CodedSection{{CodeSlug: "topic:climate", CodeID: "code-1", ID: "s1", Text: "Climate change"}}},
				ChunkID:     1,
				TotalChunks: 3,
				Next:        intPtr(2),
			},
		},
		{
			Name: "First chunk explicitly",
			Input: ChunkQuery{
				ChunkID: "1",
			},
			Expected: &ChunkResult{
				Chunk:       file.Chunk{ID: "1", Content: "Climate change impacts", Codes: []file.CodedSection{{CodeSlug: "topic:climate", CodeID: "code-1", ID: "s1", Text: "Climate change"}}},
				ChunkID:     1,
				TotalChunks: 3,
				Next:        intPtr(2),
			},
		},
		{
			Name: "Middle chunk",
			Input: ChunkQuery{
				ChunkID: "2",
			},
			Expected: &ChunkResult{
				Chunk:       file.Chunk{ID: "2", Content: "Economic policies affect", Codes: []file.CodedSection{{CodeSlug: "topic:economy", CodeID: "code-2", ID: "s1", Text: "Economic"}}},
				ChunkID:     2,
				TotalChunks: 3,
				Next:        intPtr(3),
			},
		},
		{
			Name: "Last chunk",
			Input: ChunkQuery{
				ChunkID: "3",
			},
			Expected: &ChunkResult{
				Chunk:       file.Chunk{ID: "3", Content: "Climate patterns shift", Codes: []file.CodedSection{{CodeSlug: "topic:climate", CodeID: "code-3", ID: "s1", Text: "Climate"}}},
				ChunkID:     3,
				TotalChunks: 3,
				Next:        nil,
			},
		},
		{
			Name: "Out of bounds index returns nil",
			Input: ChunkQuery{
				ChunkID: "10",
			},
			Expected: nil,
		},
		{
			Name: "Filter by searchText returns first match",
			Input: ChunkQuery{
				ChunkFilter: ChunkFilter{
					SearchText: "Climate",
				},
			},
			Expected: &ChunkResult{
				Chunk:       file.Chunk{ID: "1", Content: "Climate change impacts", Codes: []file.CodedSection{{CodeSlug: "topic:climate", CodeID: "code-1", ID: "s1", Text: "Climate change"}}},
				ChunkID:     1,
				TotalChunks: 2,
				Next:        intPtr(3),
			},
		},
		{
			Name: "Filter by searchText with specific index",
			Input: ChunkQuery{
				ChunkFilter: ChunkFilter{
					SearchText: "Climate",
				},
				ChunkID: "3",
			},
			Expected: &ChunkResult{
				Chunk:       file.Chunk{ID: "3", Content: "Climate patterns shift", Codes: []file.CodedSection{{CodeSlug: "topic:climate", CodeID: "code-3", ID: "s1", Text: "Climate"}}},
				ChunkID:     3,
				TotalChunks: 2,
				Next:        nil,
			},
		},
		{
			Name: "Filter by searchText excludes non-matching index",
			Input: ChunkQuery{
				ChunkFilter: ChunkFilter{
					SearchText: "Climate",
				},
				ChunkID: "2",
			},
			Expected: nil,
		},
		{
			Name: "Filter by code slugs with minCoverage",
			Input: ChunkQuery{
				ChunkFilter: ChunkFilter{
					CodeSlugs:   []string{"topic:climate"},
					MinCoverage: float64Ptr(0.1),
				},
			},
			Expected: &ChunkResult{
				Chunk:       file.Chunk{ID: "1", Content: "Climate change impacts", Codes: []file.CodedSection{{CodeSlug: "topic:climate", CodeID: "code-1", ID: "s1", Text: "Climate change"}}},
				ChunkID:     1,
				TotalChunks: 2,
				Next:        intPtr(3),
			},
		},
		{
			Name: "Filter by coverage",
			Input: ChunkQuery{
				ChunkFilter: ChunkFilter{
					MinCoverage: float64Ptr(0.5),
				},
			},
			Expected: &ChunkResult{
				Chunk:       file.Chunk{ID: "1", Content: "Climate change impacts", Codes: []file.CodedSection{{CodeSlug: "topic:climate", CodeID: "code-1", ID: "s1", Text: "Climate change"}}},
				ChunkID:     1,
				TotalChunks: 1,
				Next:        nil,
			},
		},
		{
			Name: "No results from filter returns nil",
			Input: ChunkQuery{
				ChunkFilter: ChunkFilter{
					SearchText: "nonexistent",
				},
			},
			Expected: nil,
		},
	}

	test_helpers.RunFunctionTests(t, tests, func(q ChunkQuery) *ChunkResult {
		return GetChunk(file1, q)
	})
}
