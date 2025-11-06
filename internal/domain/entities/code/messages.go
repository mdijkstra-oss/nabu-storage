package code

type CreateCodeData struct {
	ProjectID string `json:"project_id" validate:"required"`
	Slug      string `json:"slug" validate:"required,code_slug" normalize:"trim,lowercase"`
	Color     string `json:"color" validate:"required" normalize:"trim,lowercase"`
	Reasoning string `json:"reasoning" validate:"required"`
}

type UpdateCodeData struct {
	Slug      string `json:"slug" validate:"omitempty,code_slug" normalize:"trim,lowercase"`
	Color     string `json:"color,omitempty" validate:"omitempty" normalize:"trim,lowercase"`
	Reasoning string `json:"reasoning,omitempty" validate:"omitempty" normalize:"trim,lowercase"`
}

type DeleteCodeData struct {
	ProjectID string `json:"project_id" validate:"required"`
}

type MergeCodesData struct {
	SourceID string `json:"source_id" validate:"required"`
	TargetID string `json:"target_id" validate:"required"`
}
