package cqrs

type CommandRouter func(message *AnyMessage, publisher PublishFunc) (*AnyMessage, error)

func CombineRouters(handlers ...CommandRouter) CommandRouter {
	return func(message *AnyMessage, publisher PublishFunc) (*AnyMessage, error) {
		for _, handler := range handlers {
			ch, err := handler(message, publisher)
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

func ForPayload[P any](action Action, handler func(message *AnyMessage, payload P, publisher PublishFunc) (*AnyMessage, error)) CommandRouter {
	return func(message *AnyMessage, publisher PublishFunc) (*AnyMessage, error) {
		if message.Action != action {
			return nil, nil
		}

		var payload P
		err := EnsureValidPayload(message, &payload)
		if err != nil {
			return nil, err
		}

		return handler(message, payload, publisher)
	}
}

func LimitOnType(messageType MessageType, handler ...CommandRouter) CommandRouter {
	parentRouter := CombineRouters(handler...)

	return func(message *AnyMessage, publisher PublishFunc) (*AnyMessage, error) {
		if message.Type == messageType {
			return parentRouter(message, publisher)
		}
		return nil, nil
	}
}

func LimitOnEntity(entity AggregateType, handler ...CommandRouter) CommandRouter {
	parentRouter := CombineRouters(handler...)

	return func(message *AnyMessage, publisher PublishFunc) (*AnyMessage, error) {
		if message.AggregateType == entity {
			return parentRouter(message, publisher)
		}
		return nil, nil
	}
}

func ReadOnlyRoutes(readOnlyHandlers ...func(message *AnyMessage) error) CommandRouter {
	return func(message *AnyMessage, publishFunc PublishFunc) (*AnyMessage, error) {
		for _, handler := range readOnlyHandlers {
			err := handler(message)
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
}
