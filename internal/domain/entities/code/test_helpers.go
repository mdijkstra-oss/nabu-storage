package code

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"hermes-relay/internal/lib/utils"
)

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
		Version:  1,
		CodeData: merged,
	}
}

func CreatedCodeEvent(id, projectID, slug string) *commands.AnyMessage {
	return domain_helpers.NewDomainEvent(EntityName, id, CreatedCode, CreatedCodePayload{
		ProjectID:  projectID,
		Slug:       slug,
		Color:      "green",
		Definition: "Test definition",
	})
}
