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
		p.subscribers = utils.Filter(p.subscribers, func(sub subscription) bool {
			return sub.id != id
		})
	}
}

func (p *InMemoryPublisher) Publish(event *commands.AnyMessage) (*commands.AnyMessage, error) {
	if err := commands.ValidateMessage(event); err != nil {
		return nil, err
	}

	// Todo: Persistence happens BEFORE state changes
	// Current issue: subscribers (projections) run and update state before persistence subscriber.
	// If persistence fails, state is already changed. Need to either:
	// 1. Make persistence first subscriber (but then can't persist derived events from handlers)
	// 2. Pure event sourcing: validate → persist → then apply to projections
	// 3. Two-phase: prepare all changes, persist, then commit to projections

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
