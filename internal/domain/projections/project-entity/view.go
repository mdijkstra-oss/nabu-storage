package projectview

import (
	"hermes-relay/internal/domain/entities/project"
	documentview "hermes-relay/internal/domain/projections/document-entity"
)

type Project = project.Project

type ProjectView struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Healthy bool   `json:"healthy"`
	project.ProjectData
	Documents map[string]documentview.DocumentSummary `json:"documents"`
}

func (pv ProjectView) GetID() string {
	return pv.ID
}

type ProjectSummary struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Healthy bool   `json:"healthy"`
	project.ProjectData
}

func (ps ProjectSummary) GetID() string {
	return ps.ID
}

func ToSummary(p project.Project) ProjectSummary {
	return ProjectSummary{
		ID:          p.ID,
		Version:     p.Version,
		Healthy:     p.Healthy,
		ProjectData: p.ProjectData,
	}
}
