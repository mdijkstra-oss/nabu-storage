package cqrs

import (
	"context"
	"hermes-relay/internal/utils"
	"sync"
)

type PublishFunc func(ctx context.Context, event *Message) (*Message, error)

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

func (p *InMemoryPublisher) Publish(ctx context.Context, event *Message) (*Message, error) {
	if err := ValidateMessage(event); err != nil {
		return nil, &utils.ValidationError{Message: err.Error()}
	}

	p.mu.RLock()
	subscribers := p.subscribers
	p.mu.RUnlock()

	var resultMessages []*Message

	for _, router := range subscribers {
		result, err := router(ctx, event, p.Publish)

		if result != nil {
			utils.Must(p.Publish(ctx, result))
		}

		if err != nil {
			return nil, err
		}

		if result != nil {
			resultMessages = append(resultMessages, result)
		}
	}

	var firstResult *Message
	if len(resultMessages) > 0 {
		firstResult = resultMessages[0]
	}

	for _, msg := range resultMessages {
		_, err := p.Publish(ctx, msg)
		if err != nil {
			return firstResult, err
		}
	}

	return firstResult, nil
}
