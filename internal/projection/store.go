package projection

import (
	"fmt"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/lib/utils"
	"sync"
)

// Core types
type EventApplier interface {
	ApplyEvent(message *cqrs.AnyMessage) error
	ApplyEvents(events []cqrs.AnyMessage) error
}

type Entity interface {
	GetID() string
}

type QueryExecutor[Q, R any] func(Q) R
type FilterFunc[T Entity, Q, R any] func([]T, Q) []R

// Store
type Store[T Entity] struct {
	mu      sync.RWMutex
	data    map[string]T
	reducer cqrs.Reducer[T]
}

func NewStore[T Entity](reducer cqrs.Reducer[T]) *Store[T] {
	return &Store[T]{
		data:    make(map[string]T),
		reducer: reducer,
	}
}

func NewStoreWithDefaults[T Entity](reducer cqrs.Reducer[T], defaults []T) *Store[T] {
	store := NewStore(reducer)
	for _, item := range defaults {
		store.data[item.GetID()] = item
	}
	return store
}

// Event application
func (s *Store[T]) ApplyEvent(message *cqrs.AnyMessage) error {
	return s.ApplyEvents([]cqrs.AnyMessage{*message})
}

func (s *Store[T]) ApplyEvents(events []cqrs.AnyMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.reducer == nil {
		return fmt.Errorf("no reducer registered for store")
	}

	grouped := make(map[string][]cqrs.AnyMessage)
	for _, event := range events {
		grouped[event.AggregateID] = append(grouped[event.AggregateID], event)
	}

	for id, msgs := range grouped {
		var currentState *T
		if existing, ok := s.data[id]; ok {
			currentState = &existing
		}

		for _, event := range msgs {
			currentState = s.reducer(currentState, &event)
		}

		if currentState == nil {
			delete(s.data, id)
		} else {
			s.data[id] = *currentState
		}
	}

	return nil
}

// Direct accessors
func (s *Store[T]) GetByID(id string) *T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if state, ok := s.data[id]; ok {
		return &state
	}
	return nil
}

func (s *Store[T]) GetAll() []T {
	return Query(s, ByAll[T], EmptyQuery{})
}

// Query operations
func Query[T Entity, Q, R any](s *Store[T], filter FilterFunc[T, Q, R], q Q) []R {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return filter(utils.Values(s.data), q)
}

func QueryOne[T Entity, Q, R any](s *Store[T], filter FilterFunc[T, Q, R], q Q) *R {
	results := Query(s, filter, q)
	if len(results) == 0 {
		return nil
	}
	return &results[0]
}

// Executor binding
func BindQuery[T Entity, Q, R any](s *Store[T], filter FilterFunc[T, Q, R]) QueryExecutor[Q, []R] {
	return func(q Q) []R {
		return Query(s, filter, q)
	}
}

func BindQueryOne[T Entity, Q, R any](s *Store[T], filter FilterFunc[T, Q, R]) QueryExecutor[Q, *R] {
	return func(q Q) *R {
		return QueryOne(s, filter, q)
	}
}

func ThenMap[T Entity, Q, R any](
	filter FilterFunc[T, Q, T],
	mapper func(T) R,
) FilterFunc[T, Q, R] {
	return func(items []T, q Q) []R {
		intermediate := filter(items, q)
		return utils.Map(intermediate, mapper)
	}
}
