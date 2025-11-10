package dispatch

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/utils"
	"time"
)

func ToCreateEntityEvent[P, PE any](commandAction, eventAction commands.Action, transform ...func(*P) PE) CommandRouter {

	return func(message *commands.AnyMessage, publisher PublishFunc) (*commands.AnyMessage, error) {

		// Copies struct, so all good
		withID := *message
		withID.AggregateID = utils.NewID()
		withID.Timestamp = time.Now()

		// What is a create but an update with more fields
		return ToUpdateEntityEvent[P](commandAction, eventAction, transform...)(&withID, publisher)
	}
}

func ToUpdateEntityEvent[P, PE any](commandAction, eventAction commands.Action, transform ...func(*P) PE) CommandRouter {
	return func(message *commands.AnyMessage, publisher PublishFunc) (*commands.AnyMessage, error) {
		if message.Action != commandAction {
			return nil, nil
		}

		var payload P

		// Validate command payload
		err := commands.EnsureValidPayload(message, &payload)
		if err != nil {
			return nil, err
		}

		// If transform provided, apply and validate event payload
		if len(transform) > 0 && transform[0] != nil {
			eventPayload := transform[0](&payload)

			// Validate transformed event payload
			if err := utils.ToValidationError(utils.Validate.Struct(eventPayload)); err != nil {
				return nil, err
			}

			// Pass transformed payload to event
			return commands.ToDomainEvent[any](message, eventAction, eventPayload), nil
		}

		// No transform - use original payload (no validation needed)
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
