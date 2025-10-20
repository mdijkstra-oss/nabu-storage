package projection

import (
	"fmt"
	"hermes-relay/internal/cqrs"
	"sync"
)

// EventApplier is a type-erased interface for stores that can apply events
type EventApplier interface {
	ApplyEvent(message *cqrs.Message) error
	ApplyEvents(events []cqrs.Message) error
}

// ProjectionStore is a generic in-memory store for entities
type ProjectionStore[T any] struct {
	mu      sync.RWMutex
	data    map[string]T
	reducer cqrs.Reducer[T]
}

func NewStore[T any](reducer cqrs.Reducer[T]) *ProjectionStore[T] {
	return &ProjectionStore[T]{
		data:    make(map[string]T),
		reducer: reducer,
	}
}

func (s *ProjectionStore[T]) ApplyEvent(message *cqrs.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.reducer == nil {
		return fmt.Errorf("no reducer registered for store")
	}

	// Get current state or nil for new entities
	var currentState *T
	if existing, ok := s.data[message.AggregateID]; ok {
		currentState = &existing
	}

	newState := s.reducer(currentState, message)

	// If newState is nil, it signals deletion
	if newState == nil {
		delete(s.data, message.AggregateID)
		return nil
	}

	s.data[message.AggregateID] = *newState
	return nil
}

func (s *ProjectionStore[T]) ApplyEvents(events []cqrs.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.reducer == nil {
		return fmt.Errorf("no reducer registered for store")
	}

	// Group events by aggregate ID
	grouped := make(map[string][]cqrs.Message)
	for _, event := range events {
		grouped[event.AggregateID] = append(grouped[event.AggregateID], event)
	}

	// Apply events per aggregate
	for id, msgs := range grouped {
		var currentState *T
		if existing, ok := s.data[id]; ok {
			currentState = &existing
		}

		for _, event := range msgs {
			newState := s.reducer(currentState, &event)
			currentState = newState
		}

		// If currentState is nil, it signals deletion
		if currentState == nil {
			delete(s.data, id)
		} else {
			s.data[id] = *currentState
		}
	}
	return nil
}

func (s *ProjectionStore[T]) GetByID(id string) (*T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.data[id]
	if !ok {
		return nil, fmt.Errorf("not found: %s", id)
	}

	return &state, nil
}

func (s *ProjectionStore[T]) GetAll() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]T, 0, len(s.data))
	for _, v := range s.data {
		result = append(result, v)
	}

	return result
}

func (s *ProjectionStore[T]) DeleteByID(aggregateID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, aggregateID)
}
