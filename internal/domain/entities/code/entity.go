package code

// Todo: Make some back and forth (perhaps stage?) to get this to be great before you actually start coding
// Eg min number of entries etc.
// Probably require some fields etc

type Code struct {
	ID      string `json:"id" validate:"required"`
	Healthy bool   `json:"healthy"`
	Version int    `json:"version"`
	CodeData
}

type CodeData struct {
	ProjectID string `json:"project_id" validate:"required"`

	// eg emotion:anxiety, theme:power-dynamics
	Slug   string `json:"slug" validate:"required,min=3,max=100,code_slug" normalize:"trim,lowercase"`
	Pinned bool   `json:"pinned"`

	// Radix base color: gray, mauve, slate, sage, olive, sand, tomato, red, ruby, crimson, pink, plum,
	// purple, violet, iris, indigo, blue, cyan, teal, jade, green, grass, bronze, gold, brown, orange,
	// amber, yellow, lime, mint, sky
	Color string `json:"color" validate:"required,radix_color" normalize:"trim,lowercase"`

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

func (c Code) GetVersion() int {
	return c.Version
}

func (c Code) WithVersion(v int) any {
	c.Version = v
	return &c
}

func (c Code) WithPinned(pinned bool) any {
	c.Pinned = pinned
	return &c
}
