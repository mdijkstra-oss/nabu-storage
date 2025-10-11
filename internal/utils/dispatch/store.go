package dispatch

import (
	"encoding/json"
	"fmt"
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

func (s *Store) ApplyEvent(event Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := ApplyEvent(s.data[event.AggregateID], event)
	if err != nil {
		return err
	}
	s.data[event.AggregateID] = state
	return nil
}

func (s *Store) ApplyEvents(events []Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	grouped := make(map[string][]Message)
	for _, event := range events {
		grouped[event.AggregateID] = append(grouped[event.AggregateID], event)
	}

	for id, msgs := range grouped {
		state, err := ApplyEvents(s.data[id], msgs)
		if err != nil {
			return err
		}
		s.data[id] = state
	}
	return nil
}

func (s *Store) GetByID(id string, target interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.data[id]
	if !ok {
		return fmt.Errorf("not found: %s", id)
	}

	if err := json.Unmarshal(state, target); err != nil {
		return err
	}

	return nil
}
