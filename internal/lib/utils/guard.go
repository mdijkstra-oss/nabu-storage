package utils

import "log/slog"

// Guard wraps a function with panic recovery
func Guard(fn func()) {
	GuardWith(fn)
}

// GuardWith wraps a function with panic recovery and custom context logging
func GuardWith(fn func(), logContext ...any) {
	defer func() {
		if r := recover(); r != nil {
			args := append([]any{"panic", r}, logContext...)
			slog.Error("panic recovered", args...)
		}
	}()
	fn()
}

// GuardReturn wraps a function that returns a value with panic recovery
// Returns the zero value of T if panic occurs
func GuardReturn[T any](fn func() T) T {
	return GuardReturnWith(fn)
}

// GuardReturnWith wraps a function that returns a value with panic recovery and custom context logging
func GuardReturnWith[T any](fn func() T, logContext ...any) T {
	var result T
	defer func() {
		if r := recover(); r != nil {
			args := append([]any{"panic", r}, logContext...)
			slog.Error("panic recovered", args...)
		}
	}()
	result = fn()
	return result
}

// GuardReturnError wraps a function that returns (T, error) with panic recovery
// Returns (zero value, nil) if panic occurs
func GuardReturnError[T any](fn func() (T, error)) (T, error) {
	return GuardReturnErrorWith(fn)
}

// GuardReturnErrorWith wraps a function that returns (T, error) with panic recovery and custom context logging
func GuardReturnErrorWith[T any](fn func() (T, error), logContext ...any) (T, error) {
	var result T
	var err error
	defer func() {
		if r := recover(); r != nil {
			args := append([]any{"panic", r}, logContext...)
			slog.Error("panic recovered", args...)
		}
	}()
	result, err = fn()
	return result, err
}
