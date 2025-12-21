package project

import (
	"hermes-relay/internal/domain/entities/document"
	"time"
)

type ProjectData struct {
	Name        string    `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string    `json:"description" validate:"max=2000" normalize:"trim"`
	Pinned      bool      `json:"pinned"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Project struct {
	ID      string `json:"id" validate:"required"`
	Healthy bool   `json:"healthy"`
	Version int    `json:"version"`
	ProjectData
	Documents map[string]document.Document `json:"documents"`
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

func (p Project) WithUpdatedAt(t time.Time) any {
	p.UpdatedAt = t
	return &p
}

func (p Project) GetDocument(id string) (document.Document, bool) {
	d, exists := p.Documents[id]
	return d, exists
}
