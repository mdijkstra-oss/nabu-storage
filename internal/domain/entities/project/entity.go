package project

type Project struct {
	ID      string   `json:"id" validate:"required"`
	Name    string   `json:"name" validate:"required" normalize:"trim"`
	CodeIDs []string `json:"code_ids"`
	FileIDs []string `json:"file_ids"`
}
