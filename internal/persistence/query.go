package persistence

import (
	"encoding/json"
	"fmt"
)

func GetByID[T any](s *Store, id string) (*T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, ok := s.data[id]
	if !ok {
		return nil, fmt.Errorf("not found: %s", id)
	}

	var result T
	if err := json.Unmarshal(state, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func GetAll[T any](s *Store) ([]T, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]T, 0, len(s.data))

	for _, v := range s.data {
		var item T
		if err := json.Unmarshal(v, &item); err != nil {
			return nil, err
		}
		result = append(result, item)
	}

	return result, nil
}
