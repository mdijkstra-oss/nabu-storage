package code

type CreateCodeData struct {
	Slug      string `json:"slug" validate:"required,code_slug"`
	Color     string `json:"color" validate:"required"`
	Reasoning string `json:"reasoning" validate:"required"`
}

type UpdateCodeData struct {
	Color     string `json:"color,omitempty" validate:"omitempty"`
	Reasoning string `json:"reasoning,omitempty" validate:"omitempty"`
}
