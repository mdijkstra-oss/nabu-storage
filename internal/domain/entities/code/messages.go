package code

type CreateCodeData struct {
	ProjectID         string   `json:"project_id" validate:"required,valid_id"`
	Slug              string   `json:"slug" validate:"required,code_slug" normalize:"trim,lowercase"`
	Color             string   `json:"color" validate:"required" normalize:"trim,lowercase"`
	Definition        string   `json:"definition" validate:"required"`
	InclusionCriteria string   `json:"inclusion_criteria"`
	ExclusionCriteria string   `json:"exclusion_criteria"`
	Examples          []string `json:"examples"`
	CounterExamples   []string `json:"counter_examples"`
	Notes             string   `json:"notes"`
}

type UpdateCodeData struct {
	Slug              string   `json:"slug" validate:"omitempty,code_slug" normalize:"trim,lowercase"`
	Color             string   `json:"color,omitempty" validate:"omitempty" normalize:"trim,lowercase"`
	Definition        string   `json:"definition,omitempty"`
	InclusionCriteria string   `json:"inclusion_criteria,omitempty"`
	ExclusionCriteria string   `json:"exclusion_criteria,omitempty"`
	Examples          []string `json:"examples,omitempty"`
	CounterExamples   []string `json:"counter_examples,omitempty"`
	Notes             string   `json:"notes,omitempty"`
}

type MergeCodesData struct {
	SourceID string `json:"source_id" validate:"required,valid_id"`
	TargetID string `json:"target_id" validate:"required,valid_id"`
}
