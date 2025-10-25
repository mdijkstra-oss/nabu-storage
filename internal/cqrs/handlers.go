package cqrs

import (
	"context"
	"github.com/google/uuid"
	"time"
)

func ToCreateEntityEvent[P any](commandAction, eventAction Action) CommandRouter {

	return func(ctx context.Context, message *AnyMessage, publisher PublishFunc) (*AnyMessage, error) {

		// Copies struct, so all good
		withID := *message
		withID.AggregateID = uuid.New().String()
		withID.Timestamp = time.Now()

		// What is a create but an update with more fields
		return ToUpdateEntityEvent[P](commandAction, eventAction)(ctx, message, publisher)
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
