package projectview

import (
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/project"
	codeview "hermes-relay/internal/domain/projections/code-entity"
	"hermes-relay/internal/lib/utils"
)

func IsSlugAvailable(proj project.Project, slug string, excludeID string) bool {
	codes := utils.Values(proj.Codes)
	return codeview.IsSlugAvailable(codes, slug, excludeID)
}

func CodeExists(proj project.Project, codeID string) bool {
	return projection.ExistsInMap(proj.Codes, codeID)
}
