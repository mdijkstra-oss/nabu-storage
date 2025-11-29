package project

type CreateProjectData struct {
	Name        string `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string `json:"description" validate:"max=2000" normalize:"trim"`
	Phase       Phase  `json:"phase" validate:"omitempty,oneof=explore code validate analyze"`
}

type UpdateProjectData struct {
	Name        string `json:"name" validate:"required,max=200" normalize:"trim"`
	Description string `json:"description,omitempty" validate:"omitempty,max=2000" normalize:"trim"`
	Phase       Phase  `json:"phase,omitempty" validate:"omitempty,oneof=explore code validate analyze"`
}
