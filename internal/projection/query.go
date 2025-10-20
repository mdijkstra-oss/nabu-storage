package projection

// GetByID retrieves an entity by ID from the store
func GetByID[T any](s *ProjectionStore[T], id string) (*T, error) {
	return s.GetByID(id)
}

// GetAll retrieves all entities from the store
func GetAll[T any](s *ProjectionStore[T]) ([]T, error) {
	return s.GetAll(), nil
}
