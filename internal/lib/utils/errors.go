package utils

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"log/slog"
)

type ValidationError struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

func ToValidationError(err error) *ValidationError {
	if err == nil {
		return nil
	}

	ve := &ValidationError{
		Message: "validation failed",
		Fields:  make(map[string]string),
	}

	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		for _, fe := range validationErrs {
			ve.Fields[fe.Field()] = fe.Tag()
		}
	}

	return ve
}

type NotFoundError struct {
	Message string `json:"message"`
}

func (e *NotFoundError) Error() string {
	return e.Message
}

type ConflictError struct {
	Message string `json:"message"`
}

func (e *ConflictError) Error() string {
	return e.Message
}

type InternalError struct {
	Message string `json:"message"`
}

// Todo: Not in prod
func (e *InternalError) Error() string {
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

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func MustNotError(err error) {
	if err != nil {
		panic(err)
	}
}

// WarnErr Only for non-critical, unlikely errors (eg decoding object that is typed to always be decoded etc)
func WarnErr(err error) {
	if err == nil {
		return
	}

	slog.Warn("Error: %v", "err", err)
}
