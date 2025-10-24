package cqrs

import (
	"context"
	"github.com/google/uuid"
)

func ToCreateEntityEvent[P any](commandAction, eventAction Action) CommandRouter {
	return func(ctx context.Context, message *AnyMessage, publisher PublishFunc) (*AnyMessage, error) {
		if message.Action != commandAction {
			return nil, nil
		}

		var payload P
		err := EnsureValidPayload(message, &payload)
		if err != nil {
			return nil, err
		}

		// Copies struct, so all good
		withID := *message
		withID.AggregateID = uuid.New().String()
		return ToDomainEvent(&withID, eventAction), nil
	}
}

func ToUpdateEntityEvent[P any](commandAction, eventAction Action) CommandRouter {
	return func(ctx context.Context, message *AnyMessage, publisher PublishFunc) (*AnyMessage, error) {
		if message.Action != commandAction {
			return nil, nil
		}

		var payload P
		err := EnsureValidPayload(message, &payload)
		if err != nil {
			return nil, err
		}

		return ToDomainEvent(message, eventAction), nil
	}
}

func ToEmptyDomainEvent(commandAction, eventAction Action) CommandRouter {
	return func(ctx context.Context, message *AnyMessage, publisher PublishFunc) (*AnyMessage, error) {
		if message.Action != commandAction {
			return nil, nil
		}

		message.Payload = nil
		return ToDomainEvent(message, eventAction), nil
	}
}
