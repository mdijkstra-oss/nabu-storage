package project

type Project struct {
	ID          string   `json:"id" validate:"required"`
	Name        string   `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string   `json:"description" validate:"max=2000" normalize:"trim"`
	CodeIDs     []string `json:"code_ids"`
	FileIDs     []string `json:"file_ids"`
}

func (p Project) GetID() string {
	return p.ID
}
