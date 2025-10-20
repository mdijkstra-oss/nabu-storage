package code

type Code struct {
	ID string `json:"id" validate:"required"`
	// Color is in tailwind color system e.g. red-500 intensity on 300-700 range (inclusive)
	Color string `json:"color" validate:"required"`
}
