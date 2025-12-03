package fileview

import (
	"strconv"

	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/projections/file-entity/chunk"
	"hermes-relay/internal/lib/text-search/find"
	"hermes-relay/internal/lib/utils"
)

func QueryFiles(query projection.PaginationQuery, proj project.Project) []FileSummary {
	files := utils.Values(proj.Files)
	return utils.Map(files, ToSummary)
}

func QueryFile(query projection.IDQuery, proj project.Project) *FileSummary {
	f := projection.GetFromMap(proj.Files, query.ID)
	if f == nil {
		return nil
	}
	summary := ToSummary(*f)
	return &summary
}

func QueryChunk(query chunk.ChunkQuery, proj project.Project) *chunk.ChunkResult {
	f := projection.GetFromMap(proj.Files, query.ID)
	if f == nil {
		return nil
	}
	return chunk.GetChunk(*f, query)
}

type CodebookContent struct {
	Content string `json:"content"`
}

func QueryCodebook(query projection.EmptyQuery, proj project.Project) *CodebookContent {
	codebook := GetCodebook(proj)
	if codebook == nil || len(codebook.Chunks) == 0 {
		return nil
	}
	return &CodebookContent{Content: codebook.Chunks[0].Content}
}

func FindFileByType(proj project.Project, fileType file.FileType) *file.File {
	for _, f := range proj.Files {
		if f.Type == fileType {
			return &f
		}
	}
	return nil
}

func GetCodebook(proj project.Project) *file.File {
	return FindFileByType(proj, file.FileTypeCodebook)
}

type SearchQuery struct {
	Query            string `query:"q" validate:"required,min=3"`
	ContextSentences int    `query:"context" validate:"omitempty,min=0,max=5" default:"1"`
	SinceID          string `query:"since_id"`
	Limit            int    `query:"limit" validate:"min=1,max=100" default:"20"`
}

type SearchResult struct {
	ID       string `json:"id"`
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	ChunkID  string `json:"chunk_id"`
	Context  string `json:"context"`
	Match    string `json:"match"`
}

func (r SearchResult) GetID() string {
	return r.ID
}

func SearchProject(query SearchQuery, proj project.Project) projection.CursorResult[SearchResult] {
	files := utils.Values(proj.Files)
	allResults := utils.FlatMap(files, func(f file.File) []SearchResult {
		return searchFile(f, query)
	})
	return cursorFilterSearch(allResults, query)
}

func cursorFilterSearch(items []SearchResult, query SearchQuery) projection.CursorResult[SearchResult] {
	filtered := filterSearchBySinceID(items, query.SinceID)

	hasMore := len(filtered) > query.Limit
	if hasMore {
		filtered = filtered[:query.Limit]
	}

	nextCursor := ""
	if len(filtered) > 0 {
		nextCursor = filtered[len(filtered)-1].ID
	} else if query.SinceID != "" {
		nextCursor = query.SinceID
	}

	return projection.CursorResult[SearchResult]{
		Items:      filtered,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}

func filterSearchBySinceID(items []SearchResult, sinceID string) []SearchResult {
	if sinceID == "" {
		return items
	}
	for i, item := range items {
		if item.ID == sinceID {
			return items[i+1:]
		}
	}
	return items
}

func searchFile(f file.File, query SearchQuery) []SearchResult {
	return utils.FlatMap(f.Chunks, func(c file.Chunk) []SearchResult {
		return searchChunk(f, c, query)
	})
}

func SearchResultID(fileID, chunkID string, matchIndex int) string {
	return fileID + ":" + chunkID + ":" + strconv.Itoa(matchIndex)
}

func searchChunk(f file.File, c file.Chunk, query SearchQuery) []SearchResult {
	matches := find.FindAll(query.Query, c.Content)
	results := make([]SearchResult, len(matches))
	for i, m := range matches {
		results[i] = SearchResult{
			ID:       SearchResultID(f.ID, c.ID, i),
			FileID:   f.ID,
			FileName: f.Name,
			ChunkID:  c.ID,
			Context:  find.ExtractContext(c.Content, m.Start, m.End, query.ContextSentences),
			Match:    m.Text,
		}
	}
	return results
}
