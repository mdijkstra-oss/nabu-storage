package dispatch

import "context"

type CommandRouter func(ctx context.Context, action *Message, publisher PublishFunc) (*Message, error)
type MessageHandler func(ctx context.Context, action *Message) (*Message, error)

func MakeCombinedRouter(handlers ...CommandRouter) CommandRouter {
	return func(ctx context.Context, action *Message, publisher PublishFunc) (*Message, error) {
		for _, handler := range handlers {
			ch, err := handler(ctx, action, publisher)
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

	return func(ctx context.Context, action *Message, publisher PublishFunc) (*Message, error) {
		if action.Action == actionType {
			return parentRouter(ctx, action, publisher)
		}
		return nil, nil
	}
}

func LimitOnType(messageType MessageType, handler ...CommandRouter) CommandRouter {
	parentRouter := MakeCombinedRouter(handler...)

	return func(ctx context.Context, action *Message, publisher PublishFunc) (*Message, error) {
		if action.Type == messageType {
			return parentRouter(ctx, action, publisher)
		}
		return nil, nil
	}
}

func WithPublisher(handler CommandRouter, publisher PublishFunc) MessageHandler {
	return func(ctx context.Context, action *Message) (*Message, error) {
		return handler(ctx, action, publisher)
	}
}

func EmptyRouter() CommandRouter {
	return func(ctx context.Context, action *Message, publisher PublishFunc) (*Message, error) {
		return nil, nil
	}
}
