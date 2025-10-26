package file

import "time"

type File struct {
	BaseFile
	Content string `json:"content"`
	Chunks  []Chunk
}

type BaseFile struct {
	ID   string `json:"id"`
	Name string `json:"name" validate:"required" normalize:"trim"`

	Attributes
}

type Attributes struct {
	Title   string    `json:"title" validate:"required,min=1,max=200"`
	Summary string    `json:"summary" validate:"max=1500"` // todo: halfpage
	Time    time.Time `json:"time" validate:"omitempty,lte"`
}

type CodedSection struct {
	StartIndex int    `json:"start_index" validate:"required,gte=0"`
	EndIndex   int    `json:"end_index" validate:"required,gtfield=StartIndex"`
	CodeSlug   string `json:"code_slug" validate:"required,min=3,max=100,code_slug" normalize:"trim,lowercase"`
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
