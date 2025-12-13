package registry

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/projection"
)

func QueryProjectEvents(query projection.CursorQuery, projectID string, state *RegistryState) projection.CursorResult[commands.AnyMessage] {
	events := state.GetProjectEvents(projectID)
	return projection.CursorFilter(events, query, func(m commands.AnyMessage) string {
		return m.GetActorType()
	})
}
