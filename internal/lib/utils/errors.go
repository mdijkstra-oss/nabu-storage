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

func formatFieldError(fe validator.FieldError) string {
	field := fe.Field()
	tag := fe.Tag()
	param := fe.Param()

	switch tag {
	case "required":
		return field + " is required"
	case "min":
		return field + " must be at least " + param + " characters"
	case "max":
		return field + " must be at most " + param + " characters"
	case "code_slug":
		return field + " must match code slug format (lowercase with colon and optional dashes)"
	case "lte":
		return field + " must not be in the future"
	case "gtfield":
		return field + " must be greater than " + param
	case "gte":
		return field + " must be at least " + param
	default:
		return field + " failed validation (" + tag + ")"
	}
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
		var messages []string
		for _, fe := range validationErrs {
			ve.Fields[fe.Field()] = fe.Tag()
			messages = append(messages, formatFieldError(fe))
		}
		if len(messages) > 0 {
			ve.Message = "validation failed: " + messages[0]
			for i := 1; i < len(messages); i++ {
				ve.Message += ", " + messages[i]
			}
		}
	}

	return ve
}

func MakeValidationFieldError(field, message string) *ValidationError {
	return &ValidationError{
		Message: "validation failed",
		Fields: map[string]string{
			field: message,
		},
	}
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
