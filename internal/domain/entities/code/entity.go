package code

// Todo: Make some back and forth (perhaps stage?) to get this to be great before you actually start coding
// Eg min number of entries etc.
// Probably require some fields etc

type Code struct {
	ID      string `json:"id" validate:"required"`
	Healthy bool   `json:"healthy"`
	CodeData
}

type CodeData struct {
	ProjectID string `json:"project_id" validate:"required"`

	// eg emotion:anxiety, theme:power-dynamics
	Slug string `json:"slug" validate:"required,min=3,max=100,code_slug" normalize:"trim,lowercase"`

	// Radix colors: gray, mauve, slate, sage, olive, sand, tomato, red, ruby, crimson, pink, plum, purple, violet, iris, indigo, blue, cyan, teal, jade, green, grass, lime, mint, sky, yellow, amber, orange, brown, bronze, gold
	// Color steps 6-10 (inclusive): (lower=less prominent, higher=more prominent)
	// format [color]-[step]
	Color string `json:"color" validate:"required" normalize:"trim,lowercase"`

	// Core definition: What does this code mean? What concept/theme/pattern does it capture?
	// Be specific and explicit. 1-2 paragraphs max.
	Definition string `json:"definition" validate:"required"`

	// When to apply: Clear criteria for when this code should be used.
	// Include decision rules, context clues, and boundary conditions.
	// 1-2 paragraphs max.
	InclusionCriteria string `json:"inclusion_criteria"`

	// When NOT to apply: What might look similar but shouldn't be coded this way?
	// Helps distinguish from related codes. 1 paragraph max.
	ExclusionCriteria string `json:"exclusion_criteria"`

	// Examples: 3-5 representative text snippets that exemplify this code.
	// Include both typical and boundary cases.
	Examples []string `json:"examples"`

	// Counter-examples: 2-3 snippets that might seem like they fit but don't.
	CounterExamples []string `json:"counter_examples"`

	// Analyst notes: Your evolving thoughts, connections to theory, frequency observations
	Notes string `json:"notes"`
}

func (c Code) GetID() string {
	return c.ID
}

func (c Code) GetProjectID() string {
	return c.ProjectID
}

func (c Code) IsHealthy() bool {
	return c.Healthy
}

func (c Code) WithUnhealthy() any {
	c.Healthy = false
	return &c
}
