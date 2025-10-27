package project

type CreateProjectData struct {
	Name string `json:"name" validate:"required" normalize:"trim"`
}
