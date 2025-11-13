package file

import "time"

type File struct {
	BaseFile
	Chunks []Chunk
}

func (f File) GetID() string {
	return f.ID
}

func (f File) GetProjectID() string {
	return f.ProjectID
}

func (f *File) MarkUnhealthy() {
	f.Healthy = false
}

func (f File) IsHealthy() bool {
	return f.Healthy
}

type FileType string

const (
	FileTypeCodebook FileType = "codebook" // Formal code definitions (user-created)
	FileTypeSource   FileType = "source"   // Documents being coded (immutable)
	FileTypeMemo     FileType = "memo"     // Analytical notes (user writes)
	FileTypeContext  FileType = "context"  // LLM working memory (AI + user)
)

type BaseFile struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id" validate:"required"`
	Name        string `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string `json:"description" validate:"max=2000" normalize:"trim"`
	Healthy     bool   `json:"healthy"`
	Attributes
}

type Attributes struct {
	Title    string    `json:"title" validate:"required,min=1,max=200"`
	Summary  string    `json:"summary" validate:"max=1500"` // todo: halfpage
	Time     time.Time `json:"time" validate:"omitempty,lte"`
	Type     FileType  `json:"type" validate:"omitempty,oneof=codebook source memo context"`
	Original string    `json:"original"` // original file (eg pdf converted into File format)
	Locked   bool      `json:"locked"`   // whether file is read-only
}

type CodedSection struct {
	StartIndex int    `json:"start_index" validate:"required,gte=0"`
	EndIndex   int    `json:"end_index" validate:"required,gtfield=StartIndex"`
	CodeSlug   string `json:"code_slug" validate:"required,min=3,max=100,code_slug" normalize:"trim,lowercase"`
	CodeID     string `json:"code_id" validate:"required"`
	CodedSectionAttributes
}

type CodedSectionAttributes struct {
	Text     string `json:"text" validate:"required,min=1,max=1500"` // todo: max halfpage?
	AIReason string `json:"ai_reason" validate:"max=1500"`
	Comment  string `json:"comment" validate:"max=1500"`
}

type Chunk struct {
	ID      string         `json:"id" validate:"required"`
	Content string         `json:"content"`
	Codes   []CodedSection `json:"codes"`
}
