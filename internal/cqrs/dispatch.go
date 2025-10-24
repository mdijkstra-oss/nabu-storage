package cqrs

import (
	"context"
)

type CommandRouter func(ctx context.Context, message *AnyMessage, publisher PublishFunc) (*AnyMessage, error)

func CombineRouters(handlers ...CommandRouter) CommandRouter {
	return func(ctx context.Context, message *AnyMessage, publisher PublishFunc) (*AnyMessage, error) {
		for _, handler := range handlers {
			ch, err := handler(ctx, message, publisher)
			if err != nil {
				return nil, err
			}
			if ch != nil {
				return ch, nil
			}
		}
		return nil, nil // No handler matched
	}
}

func ForPayload[P any](action Action, handler func(ctx context.Context, message *AnyMessage, payload P, publisher PublishFunc) (*AnyMessage, error)) CommandRouter {
	return func(ctx context.Context, message *AnyMessage, publisher PublishFunc) (*AnyMessage, error) {
		if message.Action != action {
			return nil, nil
		}

		var payload P
		err := EnsureValidPayload(message, &payload)
		if err != nil {
			return nil, err
		}

		return handler(ctx, message, payload, publisher)
	}
}

func LimitOnType(messageType MessageType, handler ...CommandRouter) CommandRouter {
	parentRouter := CombineRouters(handler...)

	return func(ctx context.Context, message *AnyMessage, publisher PublishFunc) (*AnyMessage, error) {
		if message.Type == messageType {
			return parentRouter(ctx, message, publisher)
		}
		return nil, nil
	}
}

func LimitOnEntity(entity AggregateType, handler ...CommandRouter) CommandRouter {
	parentRouter := CombineRouters(handler...)

	return func(ctx context.Context, message *AnyMessage, publisher PublishFunc) (*AnyMessage, error) {
		if message.AggregateType == entity {
			return parentRouter(ctx, message, publisher)
		}
		return nil, nil
	}
}

func ReadOnlyRoutes(readOnlyHandlers ...func(message *AnyMessage) error) CommandRouter {
	return func(ctx context.Context, message *AnyMessage, publishFunc PublishFunc) (*AnyMessage, error) {
		for _, handler := range readOnlyHandlers {
			err := handler(message)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
}
