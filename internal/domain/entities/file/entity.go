package file

type File struct {
	ID      string `json:"id" validate:"required"`
	Content string `json:"content"`
	Attributes
}

type Attributes struct {
	Codes map[string][]string `json:"codes"` // update reasoning per coded section?
}
