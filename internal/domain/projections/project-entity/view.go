package projectview

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/project"
	fileview "hermes-relay/internal/domain/projections/file-entity"
)

type Project = project.Project

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
