package dispatch

import (
	"github.com/google/uuid"
	"hermes-relay/internal/cqrs/commands"
	"time"
)

func ToCreateEntityEvent[P any](commandAction, eventAction commands.Action) CommandRouter {

	return func(message *commands.AnyMessage, publisher PublishFunc) (*commands.AnyMessage, error) {

		// Copies struct, so all good
		withID := *message
		withID.AggregateID = uuid.New().String()
		withID.Timestamp = time.Now()

		// What is a create but an update with more fields
		return ToUpdateEntityEvent[P](commandAction, eventAction)(message, publisher)
	}
}

func ToUpdateEntityEvent[P any](commandAction, eventAction commands.Action) CommandRouter {
	return func(message *commands.AnyMessage, publisher PublishFunc) (*commands.AnyMessage, error) {
		if message.Action != commandAction {
			return nil, nil
		}

		var payload P
		err := commands.EnsureValidPayload(message, &payload)
		if err != nil {
			return nil, err
		}

		return commands.ToDomainEvent(message, eventAction), nil
	}
}

func ToEmptyDomainEvent(commandAction, eventAction commands.Action) CommandRouter {
	return func(message *commands.AnyMessage, publisher PublishFunc) (*commands.AnyMessage, error) {
		if message.Action != commandAction {
			return nil, nil
		}

		message.Payload = nil
		return commands.ToDomainEvent(message, eventAction), nil
	}
}
