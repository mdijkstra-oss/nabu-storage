package utils

import "github.com/google/uuid"

// NewID generates a new UUID string
func NewID() string {
	return uuid.New().String()
}

func ValidID(id string) bool {
	if id == "" {
		return false
	}
	_, err := uuid.Parse(id)
	return err == nil
}
