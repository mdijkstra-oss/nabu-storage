package utils

import "log/slog"

type ValidationError struct {
	Message string
	Fields  map[string]string
}

func (e *ValidationError) Error() string {
	if len(e.Fields) == 1 {
		for field, msg := range e.Fields {
			return field + ": " + msg
		}
	}
	return e.Message
}

func FieldError(field, message string) *ValidationError {
	return &ValidationError{
		Message: "validation failed",
		Fields:  map[string]string{field: message},
	}
}

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

func MustNotError(err error) {
	if err != nil {
		panic(err)
	}
}

func ShouldWork(err error) {
	if err != nil {
		slog.Error("tolerable error", "error", err)
	}
}
