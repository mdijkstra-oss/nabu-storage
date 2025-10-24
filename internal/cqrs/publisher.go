package cqrs

import (
	"context"
	"hermes-relay/internal/lib/utils"
	"sync"
)

type PublishFunc func(ctx context.Context, event *AnyMessage) (*AnyMessage, error)

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

func (p *InMemoryPublisher) Publish(ctx context.Context, event *AnyMessage) (*AnyMessage, error) {
	if err := ValidateMessage(event); err != nil {
		return nil, &utils.ValidationError{Message: err.Error()}
	}

	p.mu.RLock()
	subscribers := p.subscribers
	p.mu.RUnlock()

	var firstResult *AnyMessage

	for _, router := range subscribers {
		result, err := router(ctx, event, p.Publish)

		if result != nil {
			if firstResult == nil {
				firstResult = result
			}

			utils.Must(p.Publish(ctx, result))
		}

		if err != nil {
			return nil, err
		}
	}

	return firstResult, nil
}
