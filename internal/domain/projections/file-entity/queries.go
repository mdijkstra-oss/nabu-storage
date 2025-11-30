package fileview

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/projections/file-entity/chunk"
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
