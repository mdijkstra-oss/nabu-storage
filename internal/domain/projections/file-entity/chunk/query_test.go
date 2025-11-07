package chunk

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/file"
	test_helpers "hermes-relay/internal/lib/test-helpers"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	"testing"
)

func TestByChunk(t *testing.T) {
	files := []fileview.File{
		{
			BaseFile: file.BaseFile{
				ID:        "file-1",
				ProjectID: "project-1",
				Name:      "test.txt",
			},
			Content: "Test content",
			Chunks: []file.Chunk{
				{IDX: "1", Content: "First chunk", Codes: []file.CodedSection{}},
				{IDX: "2", Content: "Second chunk", Codes: []file.CodedSection{}},
				{IDX: "3", Content: "Third chunk", Codes: []file.CodedSection{}},
			},
		},
	}

	tests := []struct {
		Name     string
		Input    ChunkQuery
		Expected []ChunkResult
	}{
		{
			Name: "First chunk",
			Input: ChunkQuery{
				GetByIDQuery: projection.GetByIDQuery{ID: "file-1"},
				Index:        1,
			},
			Expected: []ChunkResult{
				{
					Chunk:       file.Chunk{IDX: "1", Content: "First chunk", Codes: []file.CodedSection{}},
					ChunkIndex:  1,
					TotalChunks: 3,
					HasNext:     true,
				},
			},
		},
		{
			Name: "Middle chunk",
			Input: ChunkQuery{
				GetByIDQuery: projection.GetByIDQuery{ID: "file-1"},
				Index:        2,
			},
			Expected: []ChunkResult{
				{
					Chunk:       file.Chunk{IDX: "2", Content: "Second chunk", Codes: []file.CodedSection{}},
					ChunkIndex:  2,
					TotalChunks: 3,
					HasNext:     true,
				},
			},
		},
		{
			Name: "Last chunk",
			Input: ChunkQuery{
				GetByIDQuery: projection.GetByIDQuery{ID: "file-1"},
				Index:        3,
			},
			Expected: []ChunkResult{
				{
					Chunk:       file.Chunk{IDX: "3", Content: "Third chunk", Codes: []file.CodedSection{}},
					ChunkIndex:  3,
					TotalChunks: 3,
					HasNext:     false,
				},
			},
		},
		{
			Name: "Out of bounds returns nil",
			Input: ChunkQuery{
				GetByIDQuery: projection.GetByIDQuery{ID: "file-1"},
				Index:        10,
			},
			Expected: nil,
		},
		{
			Name: "Zero index returns nil",
			Input: ChunkQuery{
				GetByIDQuery: projection.GetByIDQuery{ID: "file-1"},
				Index:        0,
			},
			Expected: nil,
		},
		{
			Name: "Non-existent file returns nil",
			Input: ChunkQuery{
				GetByIDQuery: projection.GetByIDQuery{ID: "nonexistent"},
				Index:        1,
			},
			Expected: nil,
		},
	}

	test_helpers.RunFunctionTests(t, tests, func(q ChunkQuery) []ChunkResult {
		return ByChunk(files, q)
	})
}
