package dispatch

import "hermes-relay/internal/cqrs/commands"

type CommandRouter func(message *commands.AnyMessage, publisher PublishFunc) (*commands.AnyMessage, error)

func CombineRouters(handlers ...CommandRouter) CommandRouter {
	return func(message *commands.AnyMessage, publisher PublishFunc) (*commands.AnyMessage, error) {
		for _, handler := range handlers {
			ch, err := handler(message, publisher)
			if err != nil {
				return nil, err
			}
			if ch != nil {
				return ch, nil
			}
		}
		return nil, nil
	}
}

func ForPayload[P any](action commands.Action, handler func(message *commands.AnyMessage, payload P, publisher PublishFunc) (*commands.AnyMessage, error)) CommandRouter {
	return func(message *commands.AnyMessage, publisher PublishFunc) (*commands.AnyMessage, error) {
		if message.Action != action {
			return nil, nil
		}

		var payload P
		err := commands.EnsureValidPayload(message, &payload)
		if err != nil {
			return nil, err
		}

		return handler(message, payload, publisher)
	}
}

func LimitOnType(messageType commands.MessageType, handler ...CommandRouter) CommandRouter {
	parentRouter := CombineRouters(handler...)

	return func(message *commands.AnyMessage, publisher PublishFunc) (*commands.AnyMessage, error) {
		if message.Type == messageType {
			return parentRouter(message, publisher)
		}
		return nil, nil
	}
}

func LimitOnEntity(entity commands.AggregateType, handler ...CommandRouter) CommandRouter {
	parentRouter := CombineRouters(handler...)

	return func(message *commands.AnyMessage, publisher PublishFunc) (*commands.AnyMessage, error) {
		if message.AggregateType == entity {
			return parentRouter(message, publisher)
		}
		return nil, nil
	}
}

func ReadOnlyRoutes(readOnlyHandlers ...func(message *commands.AnyMessage) error) CommandRouter {
	return func(message *commands.AnyMessage, publishFunc PublishFunc) (*commands.AnyMessage, error) {
		for _, handler := range readOnlyHandlers {
			err := handler(message)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
}
