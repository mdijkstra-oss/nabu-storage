package document

import "time"

type Document struct {
	ID      string `json:"id"`
	Healthy bool   `json:"healthy"`
	Version int    `json:"version"`
	DocumentData
}

func (d Document) GetID() string {
	return d.ID
}

func (d Document) GetProjectID() string {
	return d.ProjectID
}

func (d Document) IsHealthy() bool {
	return d.Healthy
}

func (d Document) WithUnhealthy() any {
	d.Healthy = false
	return &d
}

func (d Document) GetVersion() int {
	return d.Version
}

func (d Document) WithVersion(v int) any {
	d.Version = v
	return &d
}

func (d Document) WithPinned(pinned bool) any {
	d.Pinned = pinned
	return &d
}

type DocumentData struct {
	ProjectID   string    `json:"project_id" validate:"required"`
	Name        string    `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string    `json:"description" validate:"max=2000" normalize:"trim"`
	Title       string    `json:"title" validate:"omitempty,min=1,max=200"`
	Time        time.Time `json:"time" validate:"omitempty,lte"`
	Original    string    `json:"original"`
	Pinned      bool      `json:"pinned"`
	Content     []Block   `json:"content"`
}
