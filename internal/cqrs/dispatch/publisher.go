package dispatch

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/utils"
	"sync"
)

type PublishFunc func(event *commands.AnyMessage) (*commands.AnyMessage, error)

type subscription struct {
	id     string
	router CommandRouter
}

type InMemoryPublisher struct {
	subscribers []subscription
	mu          sync.RWMutex
}

func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{
		subscribers: make([]subscription, 0),
	}
}

func (p *InMemoryPublisher) Subscribe(router CommandRouter) func() {
	id := utils.NewID()

	p.mu.Lock()
	p.subscribers = append(p.subscribers, subscription{id, router})
	p.mu.Unlock()

	return func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		for i, sub := range p.subscribers {
			if sub.id == id {
				p.subscribers = append(p.subscribers[:i], p.subscribers[i+1:]...)
				break
			}
		}
	}
}

func (p *InMemoryPublisher) Publish(event *commands.AnyMessage) (*commands.AnyMessage, error) {
	if err := commands.ValidateMessage(event); err != nil {
		return nil, err
	}

	p.mu.RLock()
	subscribers := p.subscribers
	p.mu.RUnlock()

	var firstResult *commands.AnyMessage

	for _, sub := range subscribers {
		result, err := utils.GuardReturnErrorWith(func() (*commands.AnyMessage, error) {
			return sub.router(event, p.Publish)
		}, "action", event.Action, "aggregateType", event.AggregateType, "operation", "subscriber")

		if result != nil {
			if firstResult == nil {
				firstResult = result
			}

			// Cascade events should not fail the primary event.
			// Log errors but continue processing.
			// Todo: Add replay limiter
			utils.GuardWith(func() {
				utils.Should(p.Publish(result))
			}, "parentAction", event.Action, "cascadeAction", result.Action, "operation", "cascade")
		}

		if err != nil {
			return nil, err
		}
	}

	return firstResult, nil
}
