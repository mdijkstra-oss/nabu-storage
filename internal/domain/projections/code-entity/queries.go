package codeview

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/lib/utils"
)

func QueryCodes(query projection.PaginationQuery, proj project.Project) []projection.PaginationResult[code.Code] {
	codes := utils.Values(proj.Codes)
	return projection.Paginate(codes, query)
}

func QueryCode(query projection.IDQuery, proj project.Project) *code.Code {
	return projection.GetFromMap(proj.Codes, query.ID)
}

type SectionsQuery struct {
	projection.PaginationQuery
	SectionFilter
	ID string `path:"id" validate:"required,valid_id"`
}

func QuerySections(query SectionsQuery, proj project.Project) *projection.PaginationResult[CodedSectionView] {
	if _, exists := proj.Codes[query.ID]; !exists {
		return nil
	}

	sections := GetSectionsForCode(proj, query.ID)
	sections = FilterSections(sections, query.SectionFilter)

	results := projection.Paginate(sections, query.PaginationQuery)
	if len(results) == 0 {
		return nil
	}
	return &results[0]
}
