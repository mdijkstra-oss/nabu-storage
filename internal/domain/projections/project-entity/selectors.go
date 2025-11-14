package projectview

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	codeview "hermes-relay/internal/domain/projections/code-entity"
	"hermes-relay/internal/lib/utils"
)

func GetCode(proj project.Project, codeID string) *code.Code {
	return projection.GetFromMap(proj.Codes, codeID)
}

func GetAllCodes(proj project.Project) []code.Code {
	return utils.Values(proj.Codes)
}

func GetCodeBySlug(proj project.Project, slug string) *code.Code {
	codes := utils.Values(proj.Codes)
	return codeview.GetBySlug(codes, slug)
}

func IsSlugAvailable(proj project.Project, slug string, excludeID string) bool {
	codes := utils.Values(proj.Codes)
	return codeview.IsSlugAvailable(codes, slug, excludeID)
}

func CodeExists(proj project.Project, codeID string) bool {
	return projection.ExistsInMap(proj.Codes, codeID)
}

func GetFile(proj project.Project, fileID string) *file.File {
	return projection.GetFromMap(proj.Files, fileID)
}

func GetAllFiles(proj project.Project) []file.File {
	return utils.Values(proj.Files)
}

func FileExists(proj project.Project, fileID string) bool {
	return projection.ExistsInMap(proj.Files, fileID)
}
