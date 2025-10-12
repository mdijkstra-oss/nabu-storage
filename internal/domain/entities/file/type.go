package file

type File struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Attributes
}

type Attributes struct {
	Codes []Code `json:"codes"`
}

type Code struct {
	ID    string   `json:"id"`
	Texts []string `json:"texts"`
}
