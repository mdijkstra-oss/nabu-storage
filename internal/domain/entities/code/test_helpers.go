package code

import "hermes-relay/internal/lib/utils"

func BuildTestCode(id string, overrides CodeData) Code {
	defaults := CodeData{
		ProjectID:  "project-1",
		Slug:       "topic:climate",
		Color:      "green",
		Definition: "Climate topics",
	}
	merged := utils.ApplyPartialUpdate(defaults, overrides)
	return Code{
		ID:       id,
		Healthy:  true,
		CodeData: merged,
	}
}
