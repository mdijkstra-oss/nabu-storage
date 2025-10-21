package code

type Code struct {
	ID string `json:"id" validate:"required"`
	// Color is in tailwind color system e.g. red-500 intensity on 300-700 range (inclusive)
	Color string `json:"color" validate:"required"`
	// The reasoning for determining whether this code applies to a given piece of text.
	// Claude: Be thorough, add samples, perhaps exclusions, in a way that if you see this again later it makes it possible for you to spot all the proper instances again.
	// Claude: Do not literally write out the whole list you found, limit it to a paragraph or 3 max.
	Reasoning string `json:"reasoning" validate:"required"`
}
