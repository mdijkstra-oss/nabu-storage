package projectview

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/project"
	fileview "hermes-relay/internal/domain/projections/file-entity"
)

type Project = project.Project

type ProjectView struct {
	ID      string `json:"id"`
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
		Healthy:     p.Healthy,
		ProjectData: p.ProjectData,
		Codes:       p.Codes,
		Files:       fileSummaries,
	}
}
