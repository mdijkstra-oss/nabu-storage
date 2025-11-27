package file

import (
	"hermes-relay/internal/cqrs/commands"
	"time"
)

type File struct {
	ID      string `json:"id"`
	Healthy bool   `json:"healthy"`
	FileData
	Chunks []Chunk `json:"chunks"`
}

func (f File) GetID() string {
	return f.ID
}

func (f File) GetProjectID() string {
	return f.ProjectID
}

func (f File) IsHealthy() bool {
	return f.Healthy
}

func (f File) WithUnhealthy() any {
	f.Healthy = false
	return &f
}

type FileType string

const (
	FileTypeCodebook FileType = "codebook" // Formal code definitions (user-created)
	FileTypeSource   FileType = "source"   // Documents being coded (immutable)
	FileTypeMemo     FileType = "memo"     // Analytical notes (user writes)
	FileTypeContext  FileType = "context"  // LLM working memory (AI + user)
)

type FileData struct {
	ProjectID   string    `json:"project_id" validate:"required"`
	Name        string    `json:"name" validate:"required,max=200" normalize:"trim"` // filename
	Description string    `json:"description" validate:"max=2000" normalize:"trim"`
	Title       string    `json:"title" validate:"omitempty,min=1,max=200"` // pretty name perhaps from context
	Summary     string    `json:"summary" validate:"max=1500"`              // todo: halfpage
	Time        time.Time `json:"time" validate:"omitempty,lte"`
	Type        FileType  `json:"type" validate:"omitempty,oneof=codebook source memo context"`
	Original    string    `json:"original"` // original file (eg pdf converted into File format)
	Locked      bool      `json:"locked"`   // whether file is read-only
}

type CodedSection struct {
	ID        string         `json:"id" validate:"omitempty,valid_id"`
	CodeSlug  string         `json:"code_slug" validate:"required,min=3,max=100,code_slug" normalize:"trim,lowercase"`
	CodeID    string         `json:"code_id" validate:"required"`
	Text      string         `json:"text" validate:"required,min=1,max=1500"`
	Reason    string         `json:"reason" validate:"max=1500"`
	LastActor commands.Actor `json:"last_actor"`
}

type Chunk struct {
	ID      string         `json:"id" validate:"required"`
	Content string         `json:"content"`
	Codes   []CodedSection `json:"codes"`
}
