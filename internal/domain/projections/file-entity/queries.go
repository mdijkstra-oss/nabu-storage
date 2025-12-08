package fileview

import (
	"strconv"

	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/text-search/find"
	"hermes-relay/internal/lib/utils"
)

func QueryFiles(query projection.PaginationQuery, proj project.Project) []FileSummary {
	files := utils.Values(proj.Files)
	return utils.Map(files, ToSummary)
}

func QueryFile(query projection.IDQuery, proj project.Project) *File {
	f := projection.GetFromMap(proj.Files, query.ID)
	if f == nil {
		return nil
	}
	return f
}

type CodebookContent struct {
	Content string `json:"content"`
}

func QueryCodebook(query projection.EmptyQuery, proj project.Project) *CodebookContent {
	codebook := GetCodebook(proj)
	if codebook == nil {
		return nil
	}
	return &CodebookContent{Content: codebook.Content}
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
	matches := find.FindAll(query.Query, f.Content)
	results := make([]SearchResult, len(matches))
	for i, m := range matches {
		results[i] = SearchResult{
			ID:       SearchResultID(f.ID, i),
			FileID:   f.ID,
			FileName: f.Name,
			Context:  find.ExtractContext(f.Content, m.Start, m.End, query.ContextSentences),
			Match:    m.Text,
		}
	}
	return results
}

func SearchResultID(fileID string, matchIndex int) string {
	return fileID + ":" + strconv.Itoa(matchIndex)
}
