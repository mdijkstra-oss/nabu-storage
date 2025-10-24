package file

type File struct {
	BaseFile
	Content string `json:"content"`
	Chunks  []Chunk
}

type BaseFile struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
	Name    string `json:"name" validate:"required"`
	Title   string `json:"title"`
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

func ToBaseFile(file *File) BaseFile {
	return file.BaseFile
}
