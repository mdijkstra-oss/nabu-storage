package utils_test

import (
	"errors"
	"testing"

	"hermes-relay/internal/lib/utils"
)

func TestShouldWork(t *testing.T) {
	tests := []struct {
		Name string
		Err  error
	}{
		{Name: "nil error", Err: nil},
		{Name: "with error", Err: errors.New("test")},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			utils.ShouldWork(tt.Err)
		})
	}
}

func TestMustNotError(t *testing.T) {
	tests := []struct {
		Name        string
		Err         error
		ShouldPanic bool
	}{
		{Name: "nil error does not panic", Err: nil, ShouldPanic: false},
		{Name: "error panics", Err: errors.New("fail"), ShouldPanic: true},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			panicked := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
					}
				}()
				utils.MustNotError(tt.Err)
			}()
			if panicked != tt.ShouldPanic {
				t.Errorf("expected panic=%v, got panic=%v", tt.ShouldPanic, panicked)
			}
		})
	}
}

func TestFieldError(t *testing.T) {
	tests := []struct {
		Name          string
		Field         string
		Message       string
		ExpectMessage string
		ExpectField   string
	}{
		{
			Name:          "creates validation error",
			Field:         "name",
			Message:       "required",
			ExpectMessage: "validation failed",
			ExpectField:   "required",
		},
		{
			Name:          "different field",
			Field:         "email",
			Message:       "invalid format",
			ExpectMessage: "validation failed",
			ExpectField:   "invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := utils.FieldError(tt.Field, tt.Message)
			if err.Message != tt.ExpectMessage {
				t.Errorf("expected message %q, got %q", tt.ExpectMessage, err.Message)
			}
			if err.Fields[tt.Field] != tt.ExpectField {
				t.Errorf("expected field %q=%q, got %v", tt.Field, tt.ExpectField, err.Fields)
			}
		})
	}
}
