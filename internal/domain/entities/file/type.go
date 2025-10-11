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
	Name  string   `json:"name"`
	Texts []string `json:"texts"`
}
