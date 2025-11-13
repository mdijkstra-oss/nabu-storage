package project

import (
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
)

type Project struct {
	ID          string                  `json:"id" validate:"required"`
	Name        string                  `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string                  `json:"description" validate:"max=2000" normalize:"trim"`
	Codes       map[string]code.Code    `json:"codes"`
	Files       map[string]file.File    `json:"files"`
	Healthy     bool                    `json:"healthy"`
}

func (p Project) GetID() string {
	return p.ID
}

func (p *Project) MarkUnhealthy() {
	p.Healthy = false
}

func (p Project) IsHealthy() bool {
	return p.Healthy
}
