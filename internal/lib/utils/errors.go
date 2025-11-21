package utils

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log/slog"
)

type ValidationError struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func (e *ValidationError) Error() string {
	if e.Message != "validation failed" || len(e.Fields) == 0 {
		return e.Message
	}

	var messages []string
	for field, msg := range e.Fields {
		messages = append(messages, formatFieldMessage(field, msg))
	}

	if len(messages) == 0 {
		return e.Message
	}

	result := "validation failed: " + messages[0]
	for i := 1; i < len(messages); i++ {
		result += ", " + messages[i]
	}
	return result
}

func formatFieldMessage(field, message string) string {
	return field + " " + message
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

func FieldError(field, message string) *ValidationError {
	return &ValidationError{
		Message: "validation failed",
		Fields: map[string]string{
			field: message,
		},
	}
}

func FieldNotFound(field string) *ValidationError {
	return FieldError(field, "not found")
}

func FieldInUse(field string) *ValidationError {
	return FieldError(field, "already in use")
}

func ArrayItemErrors(field string, failures map[int]string) *ValidationError {
	fields := make(map[string]string)
	var messages []string

	for index, reason := range failures {
		key := fmt.Sprintf("%s[%d]", field, index)
		fields[key] = reason
		messages = append(messages, fmt.Sprintf("%s[%d] %s", field, index, reason))
	}

	message := fmt.Sprintf("validation failed: %d items failed", len(failures))
	if len(messages) > 0 {
		message = "validation failed: " + messages[0]
		for i := 1; i < len(messages); i++ {
			message += ", " + messages[i]
		}
	}

	return &ValidationError{
		Message: message,
		Fields:  fields,
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

func Should[T any](value T, err error) T {
	if err != nil {
		slog.Warn("Should detected error", "error", err)
	}
	return value
}

func ShouldWork(err error) {
	if err == nil {
		return
	}

	slog.Warn("Error: %v", "err", err)
}
