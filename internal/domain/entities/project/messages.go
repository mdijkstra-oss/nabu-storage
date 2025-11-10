package project

type CreateProjectData struct {
	Name        string `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string `json:"description" validate:"max=2000" normalize:"trim"`
}

type UpdateProjectData struct {
	Name        string `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string `json:"description" validate:"max=2000" normalize:"trim"`
}

type AddedFileToProjectData struct {
	FileID    string `json:"file_id" validate:"required"`
	ProjectID string `json:"project_id" validate:"required"`
}

type AddedCodeToProjectData struct {
	CodeID    string `json:"code_id" validate:"required"`
	ProjectID string `json:"project_id" validate:"required"`
}

type RemovedCodeFromProjectData struct {
	CodeID    string `json:"code_id" validate:"required"`
	ProjectID string `json:"project_id" validate:"required"`
}
