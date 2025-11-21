package code

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
