package dispatch

import "context"

type CommandRouter func(ctx context.Context, message *Message, publisher PublishFunc) (*Message, error)
type MessageHandler func(ctx context.Context, message *Message) (*Message, error)

func MakeCombinedRouter(handlers ...CommandRouter) CommandRouter {
	return func(ctx context.Context, message *Message, publisher PublishFunc) (*Message, error) {
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

func LimitOnAction(actionType string, handler ...CommandRouter) CommandRouter {
	parentRouter := MakeCombinedRouter(handler...)

	return func(ctx context.Context, message *Message, publisher PublishFunc) (*Message, error) {
		if message.Action == actionType {
			return parentRouter(ctx, message, publisher)
		}
		return nil, nil
	}
}

func LimitOnType(messageType MessageType, handler ...CommandRouter) CommandRouter {
	parentRouter := MakeCombinedRouter(handler...)

	return func(ctx context.Context, message *Message, publisher PublishFunc) (*Message, error) {
		if message.Type == messageType {
			return parentRouter(ctx, message, publisher)
		}
		return nil, nil
	}
}

func WithPublisher(handler CommandRouter, publisher PublishFunc) MessageHandler {
	return func(ctx context.Context, message *Message) (*Message, error) {
		return handler(ctx, message, publisher)
	}
}

func EmptyRouter() CommandRouter {
	return func(ctx context.Context, message *Message, publisher PublishFunc) (*Message, error) {
		return nil, nil
	}
}
