package persistence

import (
	"hermes-relay/internal/commands"
	commandEvents "hermes-relay/internal/commands/events"
	"sync"
)

type Store struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewStore() *Store {
	return &Store{
		data: make(map[string][]byte),
	}
}

func (s *Store) ApplyEvent(event commands.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := commandEvents.ApplyEvent(s.data[event.AggregateID], event)
	if err != nil {
		return err
	}
	s.data[event.AggregateID] = state
	return nil
}

func (s *Store) ApplyEvents(events []commands.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	grouped := make(map[string][]commands.Message)
	for _, event := range events {
		grouped[event.AggregateID] = append(grouped[event.AggregateID], event)
	}

	for id, msgs := range grouped {
		state, err := commandEvents.ApplyEvents(s.data[id], msgs)
		if err != nil {
			return err
		}
		s.data[id] = state
	}
	return nil
}

