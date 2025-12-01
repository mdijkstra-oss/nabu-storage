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
