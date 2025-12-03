package fileview

import (
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"testing"
)

func TestSearchProject(t *testing.T) {
	tests := []struct {
		name       string
		query      SearchQuery
		proj       project.Project
		wantCount  int
		wantFileID string
		wantMore   bool
	}{
		{
			name:  "finds match in single file",
			query: SearchQuery{Query: "hello world", ContextSentences: 1, Limit: 20},
			proj: project.Project{
				Files: map[string]file.File{
					"file-1": {
						ID:       "file-1",
						FileData: file.FileData{Name: "test.md"},
						Chunks: []file.Chunk{
							{ID: "1", Content: "This contains hello world in it."},
						},
					},
				},
			},
			wantCount:  1,
			wantFileID: "file-1",
			wantMore:   false,
		},
		{
			name:  "finds matches across multiple files",
			query: SearchQuery{Query: "important", ContextSentences: 1, Limit: 20},
			proj: project.Project{
				Files: map[string]file.File{
					"file-1": {
						ID:       "file-1",
						FileData: file.FileData{Name: "first.md"},
						Chunks: []file.Chunk{
							{ID: "1", Content: "This is important information."},
						},
					},
					"file-2": {
						ID:       "file-2",
						FileData: file.FileData{Name: "second.md"},
						Chunks: []file.Chunk{
							{ID: "1", Content: "Also important here."},
						},
					},
				},
			},
			wantCount: 2,
			wantMore:  false,
		},
		{
			name:  "finds multiple matches in same chunk",
			query: SearchQuery{Query: "the cat", ContextSentences: 1, Limit: 20},
			proj: project.Project{
				Files: map[string]file.File{
					"file-1": {
						ID:       "file-1",
						FileData: file.FileData{Name: "cats.md"},
						Chunks: []file.Chunk{
							{ID: "1", Content: "The cat sat. The cat ran. The cat slept."},
						},
					},
				},
			},
			wantCount: 3,
			wantMore:  false,
		},
		{
			name:  "no matches returns empty",
			query: SearchQuery{Query: "nonexistent phrase", ContextSentences: 1, Limit: 20},
			proj: project.Project{
				Files: map[string]file.File{
					"file-1": {
						ID:       "file-1",
						FileData: file.FileData{Name: "test.md"},
						Chunks: []file.Chunk{
							{ID: "1", Content: "This has different content."},
						},
					},
				},
			},
			wantCount: 0,
			wantMore:  false,
		},
		{
			name:  "empty project returns empty",
			query: SearchQuery{Query: "anything", ContextSentences: 1, Limit: 20},
			proj: project.Project{
				Files: map[string]file.File{},
			},
			wantCount: 0,
			wantMore:  false,
		},
		{
			name:  "respects limit and sets has_more",
			query: SearchQuery{Query: "the cat", ContextSentences: 1, Limit: 2},
			proj: project.Project{
				Files: map[string]file.File{
					"file-1": {
						ID:       "file-1",
						FileData: file.FileData{Name: "cats.md"},
						Chunks: []file.Chunk{
							{ID: "1", Content: "The cat sat. The cat ran. The cat slept."},
						},
					},
				},
			},
			wantCount: 2,
			wantMore:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SearchProject(tt.query, tt.proj)

			if len(result.Items) != tt.wantCount {
				t.Errorf("SearchProject() got %d results, want %d", len(result.Items), tt.wantCount)
			}

			if result.HasMore != tt.wantMore {
				t.Errorf("SearchProject() HasMore = %v, want %v", result.HasMore, tt.wantMore)
			}

			if tt.wantFileID != "" && len(result.Items) > 0 {
				if result.Items[0].FileID != tt.wantFileID {
					t.Errorf("SearchProject() first result FileID = %q, want %q", result.Items[0].FileID, tt.wantFileID)
				}
			}
		})
	}
}

func TestSearchPagination(t *testing.T) {
	proj := project.Project{
		Files: map[string]file.File{
			"file-1": {
				ID:       "file-1",
				FileData: file.FileData{Name: "cats.md"},
				Chunks: []file.Chunk{
					{ID: "1", Content: "The cat sat. The cat ran. The cat slept."},
				},
			},
		},
	}

	firstPage := SearchProject(SearchQuery{Query: "the cat", Limit: 2}, proj)

	if len(firstPage.Items) != 2 {
		t.Fatalf("first page: expected 2 items, got %d", len(firstPage.Items))
	}
	if !firstPage.HasMore {
		t.Error("first page: expected HasMore=true")
	}
	if firstPage.NextCursor == "" {
		t.Error("first page: expected non-empty NextCursor")
	}

	secondPage := SearchProject(SearchQuery{Query: "the cat", Limit: 2, SinceID: firstPage.NextCursor}, proj)

	if len(secondPage.Items) != 1 {
		t.Fatalf("second page: expected 1 item, got %d", len(secondPage.Items))
	}
	if secondPage.HasMore {
		t.Error("second page: expected HasMore=false")
	}
}

func TestSearchResultContainsContext(t *testing.T) {
	query := SearchQuery{Query: "target word", ContextSentences: 1, Limit: 20}
	proj := project.Project{
		Files: map[string]file.File{
			"file-1": {
				ID:       "file-1",
				FileData: file.FileData{Name: "test.md"},
				Chunks: []file.Chunk{
					{ID: "1", Content: "Before sentence. The target word is here. After sentence."},
				},
			},
		},
	}

	result := SearchProject(query, proj)

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Items))
	}

	item := result.Items[0]

	if item.ID == "" {
		t.Error("ID should not be empty")
	}

	if item.Match == "" {
		t.Error("Match should not be empty")
	}

	if item.Context == "" {
		t.Error("Context should not be empty")
	}

	if item.FileName != "test.md" {
		t.Errorf("FileName = %q, want %q", item.FileName, "test.md")
	}

	if item.ChunkID != "1" {
		t.Errorf("ChunkID = %q, want %q", item.ChunkID, "1")
	}
}
