package projectview

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/project"
	fileview "hermes-relay/internal/domain/projections/file-entity"
)

type Project = project.Project

type ProjectView struct {
	ID          string                         `json:"id"`
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	Codes       map[string]code.Code           `json:"codes"`
	Files       map[string]fileview.FileSummary `json:"files"`
	Healthy     bool                           `json:"healthy"`
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
		Name:        p.Name,
		Description: p.Description,
		Codes:       p.Codes,
		Files:       fileSummaries,
		Healthy:     p.Healthy,
	}
}
