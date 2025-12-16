package project

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
)

type Phase string

const (
	PhaseExplore  Phase = "explore"
	PhaseCode     Phase = "code"
	PhaseValidate Phase = "validate"
	PhaseAnalyze  Phase = "analyze"
)

type ProjectData struct {
	Name        string `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string `json:"description" validate:"max=2000" normalize:"trim"`
	Phase       Phase  `json:"phase" validate:"oneof=explore code validate analyze"`
	Pinned      bool   `json:"pinned"`
}

type Project struct {
	ID      string `json:"id" validate:"required"`
	Healthy bool   `json:"healthy"`
	Version int    `json:"version"`
	ProjectData
	Codes map[string]code.Code `json:"codes"`
	Files map[string]file.File `json:"files"`
}

func (p Project) GetID() string {
	return p.ID
}

func (p Project) IsHealthy() bool {
	return p.Healthy
}

func (p Project) WithUnhealthy() any {
	p.Healthy = false
	return &p
}

func (p Project) IsPatchable() bool {
	return true
}

func (p Project) GetVersion() int {
	return p.Version
}

func (p Project) WithVersion(v int) any {
	p.Version = v
	return &p
}

func (p Project) WithPinned(pinned bool) any {
	p.Pinned = pinned
	return &p
}

func (p Project) GetCode(id string) (code.Code, bool) {
	c, exists := p.Codes[id]
	return c, exists
}

func (p Project) GetFile(id string) (file.File, bool) {
	f, exists := p.Files[id]
	return f, exists
}
