package dispatch

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/utils"
	"sync"
)

type PublishFunc func(event *commands.AnyMessage) (*commands.AnyMessage, error)

type InMemoryPublisher struct {
	subscribers []CommandRouter
	mu          sync.RWMutex
}

func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{
		subscribers: make([]CommandRouter, 0),
	}
}

func (p *InMemoryPublisher) Subscribe(router CommandRouter) {
	p.subscribers = append(p.subscribers, router)
}

func (p *InMemoryPublisher) Publish(event *commands.AnyMessage) (*commands.AnyMessage, error) {
	if err := commands.ValidateMessage(event); err != nil {
		return nil, err
	}

	p.mu.RLock()
	subscribers := p.subscribers
	p.mu.RUnlock()

	var firstResult *commands.AnyMessage

	for _, router := range subscribers {
		result, err := router(event, p.Publish)

		if result != nil {
			if firstResult == nil {
				firstResult = result
			}

			utils.Must(p.Publish(result))
		}

		if err != nil {
			return nil, err
		}
	}

	return firstResult, nil
}
