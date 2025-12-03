package projectview

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	fileview "hermes-relay/internal/domain/projections/file-entity"
	"hermes-relay/internal/lib/utils"
)

type Project = project.Project

// LLM: if you modify ProjectView, also update ProjectArrayView below.
type ProjectView struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Healthy bool   `json:"healthy"`
	project.ProjectData
	Codes map[string]code.Code            `json:"codes"`
	Files map[string]fileview.FileSummary `json:"files"`
}

func (pv ProjectView) GetID() string {
	return pv.ID
}

func ToView(p project.Project) ProjectView {
	fileSummaries := make(map[string]fileview.FileSummary)
	for id, f := range p.Files {
		fileSummaries[id] = fileview.ToSummary(f)
	}

	return ProjectView{
		ID:          p.ID,
		Version:     p.Version,
		Healthy:     p.Healthy,
		ProjectData: p.ProjectData,
		Codes:       p.Codes,
		Files:       fileSummaries,
	}
}

// Dict children: fast lookup during frontend searching and CQRS event applying.
// Array children: SQL querying and JSON schema generation.
type ProjectArrayView struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Healthy bool   `json:"healthy"`
	project.ProjectData
	Codes []code.Code `json:"codes"`
	Files []file.File `json:"files"`
}

func ToArrayView(p project.Project) ProjectArrayView {
	return ProjectArrayView{
		ID:          p.ID,
		Version:     p.Version,
		Healthy:     p.Healthy,
		ProjectData: p.ProjectData,
		Codes:       utils.Values(p.Codes),
		Files:       utils.Values(p.Files),
	}
}
