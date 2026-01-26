package utils

import "log/slog"

func GuardWith(fn func(), logContext ...any) {
	defer func() {
		if r := recover(); r != nil {
			args := append([]any{"panic", r}, logContext...)
			slog.Error("panic recovered", args...)
		}
	}()
	fn()
}
