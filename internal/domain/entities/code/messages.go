package code

type CreateCodeData struct {
	Color string `json:"color" validate:"required"`
}

type UpdateCodeData struct {
	Color string `json:"color,omitempty" validate:"omitempty"`
}
