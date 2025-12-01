package file

import (
	"hermes-relay/internal/cqrs/commands"
	"time"
)

type File struct {
	ID      string `json:"id"`
	Healthy bool   `json:"healthy"`
	Version int    `json:"version"`
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

func (f File) GetVersion() int {
	return f.Version
}

func (f File) WithVersion(v int) any {
	f.Version = v
	return &f
}

type FileType string

const (
	FileTypeCorpus   FileType = "corpus"   // Documents being coded (immutable)
	FileTypeCodebook FileType = "codebook" // Coding guidelines (editable)
	FileTypeMemo     FileType = "memo"     // Analytical notes (editable)
	FileTypeLLMMemo  FileType = "llm-memo" // LLM analytical notes (editable)
)

func (t FileType) IsLocked() bool {
	return t == FileTypeCorpus
}

// IsChunked: INVARIANT - chunked files are immutable after creation.
// Chunk IDs become permanent references for CodedSections. Re-chunking would
// invalidate all existing coding work. Content edits are blocked for chunked files.
func (t FileType) IsChunked() bool {
	return t == FileTypeCorpus
}

func (t FileType) IsSingleton() bool {
	return t == FileTypeCodebook || t == FileTypeLLMMemo
}

type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

type FileData struct {
	ProjectID   string    `json:"project_id" validate:"required"`
	Name        string    `json:"name" validate:"required,max=200" normalize:"trim"` // filename
	Description string    `json:"description" validate:"max=2000" normalize:"trim"`
	Title       string    `json:"title" validate:"omitempty,min=1,max=200"` // pretty name perhaps from context
	Summary     string    `json:"summary" validate:"max=1500"`              // todo: halfpage
	Time        time.Time `json:"time" validate:"omitempty,lte"`
	Type        FileType  `json:"type" validate:"omitempty,oneof=corpus codebook memo llm-memo"`
	Original    string    `json:"original"` // original file (eg pdf converted into File format)
	Locked      bool      `json:"locked"`   // whether file is read-only
}

type CodedSection struct {
	ID         string         `json:"id" validate:"omitempty,valid_id"`
	CodeID     string         `json:"code_id" validate:"required,valid_id"`
	Text       string         `json:"text" validate:"required,min=1,max=1500"`
	Reason     string         `json:"reason" validate:"max=1500"`
	Confidence Confidence     `json:"confidence" validate:"required,oneof=high medium low"`
	LastActor  commands.Actor `json:"last_actor"`
}

type Chunk struct {
	ID      string         `json:"id" validate:"required"`
	Content string         `json:"content"`
	Codes   []CodedSection `json:"codes"`
}
