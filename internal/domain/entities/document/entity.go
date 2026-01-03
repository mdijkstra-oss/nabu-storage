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

func (d Document) WithUpdatedAt(t time.Time) any {
	d.UpdatedAt = t
	return &d
}

type DocumentData struct {
	ProjectID   string       `json:"project_id" validate:"required"`
	Name        string       `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string       `json:"description" validate:"max=2000" normalize:"trim"`
	Title       string       `json:"title" validate:"omitempty,min=1,max=200"`
	Time        time.Time    `json:"time" validate:"omitempty,lte"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Original    string       `json:"original"`
	Pinned      bool         `json:"pinned"`
	Tags        []string     `json:"tags"`
	Content     []Block      `json:"content"`
	Annotations []Annotation `json:"annotations"`
}

type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

type Annotation struct {
	ID      string         `json:"id"`
	Text    string         `json:"text" validate:"required"`
	Actor   string         `json:"actor" validate:"required"`
	Color   string         `json:"color" validate:"required,radix_color"`
	Reason  string         `json:"reason,omitempty"`
	Payload *CodingPayload `json:"payload,omitempty"`
}

type CodingPayload struct {
	Type       string     `json:"type" validate:"required,eq=coding"`
	CodeID     string     `json:"code_id" validate:"required,code_id" normalize:"code_id"`
	Confidence Confidence `json:"confidence" validate:"required,oneof=high medium low"`
}
