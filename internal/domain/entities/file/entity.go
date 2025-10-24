package file

type File struct {
	ID      string  `json:"id" validate:"required"`
	Content string  `json:"content"`
	Chunks  []Chunk `json:"chunks" validate:"required,min=1,dive,required"`
}

type CodedSection struct {
	StartIndex int    `json:"start_index" validate:"required"`
	EndIndex   int    `json:"end_index" validate:"required"`
	CodeSlug   string `json:"code_slug" validate:"required"`
	Text       string `json:"text"`
	AIReason   string `json:"ai_reason"`
	Comment    string `json:"comment"`
}

type Chunk struct {
	ID      string         `json:"id" validate:"required"`
	Content string         `json:"content"`
	Codes   []CodedSection `json:"codes"`
}
