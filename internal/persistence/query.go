package persistence

// GetByID retrieves an entity by ID from the store
func GetByID[T any](s *Store[T], id string) (*T, error) {
	return s.GetByID(id)
}

// GetAll retrieves all entities from the store
func GetAll[T any](s *Store[T]) ([]T, error) {
	return s.GetAll(), nil
}
