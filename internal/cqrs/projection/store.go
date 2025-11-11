package projection

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/lib/utils"
	"sync"
)

type EventApplier interface {
	ApplyEvent(message *commands.AnyMessage) error
	ApplyEvents(events []commands.AnyMessage) error
}

type Entity interface {
	GetID() string
}

type QueryExecutor[Q, R any] func(Q) R
type FilterFunc[T Entity, Q, R any] func([]T, Q) []R

type Store[T Entity] struct {
	mu      sync.RWMutex
	data    map[string]T
	reducer Reducer[T]
}

func NewStore[T Entity](reducer Reducer[T]) *Store[T] {
	return &Store[T]{
		data:    make(map[string]T),
		reducer: reducer,
	}
}

func NewStoreWithDefaults[T Entity](reducer Reducer[T], defaults []T) *Store[T] {
	store := NewStore(reducer)
	for _, item := range defaults {
		store.data[item.GetID()] = item
	}
	return store
}

func (s *Store[T]) ApplyEvent(message *commands.AnyMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, exists := s.data[message.AggregateID]

	// Important that events that listen to creation are called first for references
	if exists {
		s.applyToEntity(current, message)
	} else if commands.IsCreatedEvent(message.Action) {
		if newState := s.reducer(nil, message); newState != nil {
			s.createEntity(newState)
		}
	}

	s.applyAcrossEntities(message)
}

func (s *Store[T]) applyToEntity(current T, message *commands.AnyMessage) {
	currentPtr := &current
	newState := s.reducer(currentPtr, message)

	if newState == nil {
		delete(s.data, message.AggregateID)
	} else {
		entityID := (*newState).GetID()
		s.data[entityID] = *newState
	}
}

func (s *Store[T]) createEntity(newState *T) {
	entityID := (*newState).GetID()
	s.data[entityID] = *newState
}

func (s *Store[T]) applyAcrossEntities(message *commands.AnyMessage) {
	for _, entity := range s.data {
		currentPtr := &entity
		newState := s.reducer(currentPtr, message)
		// Across events never delete!
		if newState != nil {
			entityID := (*newState).GetID()
			s.data[entityID] = *newState
		}
	}
}

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
