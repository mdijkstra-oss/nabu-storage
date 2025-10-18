package utils

import (
	"errors"
	"fmt"
	"log/slog"
)

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}

func IsValidationError(err error) bool {
	var validationError *ValidationError
	return errors.As(err, &validationError)
}

func IsNotFoundError(err error) bool {
	var notFoundError *NotFoundError
	return errors.As(err, &notFoundError)
}

func IsConflictError(err error) bool {
	var conflictError *ConflictError
	return errors.As(err, &conflictError)
}

func GetErrorType(err error) string {
	switch {
	case IsValidationError(err):
		return "validation_error"
	case IsNotFoundError(err):
		return "not_found"
	case IsConflictError(err):
		return "conflict"
	default:
		return "internal_error"
	}
}

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func Must2[T any, D any](value T, secondV D, err error) (T, D) {
	if err != nil {
		panic(err)
	}
	return value, secondV
}

func MustNotError(err error) {
	if err != nil {
		panic(err)
	}
}

// WarnErr Only for non critical, unlikely errors (eg decoding object that is typed to always be decoded etc)
func WarnErr(err error) {
	if err == nil {
		return
	}
	
	slog.Warn("Error: %v", "err", err)
}
