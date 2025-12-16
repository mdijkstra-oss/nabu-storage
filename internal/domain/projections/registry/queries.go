package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
)

func QueryProjectEvents(query projection.EmptyQuery, projectID string, state *RegistryState) []commands.AnyMessage {
	events := state.GetProjectEvents(projectID)
	// TODO: paginate on this query
	return events
}
