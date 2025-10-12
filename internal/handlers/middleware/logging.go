package middleware

import (
	"context"
	commands2 "hermes-relay/internal/commands"
	"log/slog"
)

func WithLogging(level slog.Level) commands2.CommandRouter {
	return func(ctx context.Context, action *commands2.Message, publisher commands2.PublishFunc) (*commands2.Message, error) {
		slog.Log(context.Background(), level, "message received", "message", action)
		return nil, nil
	}
}
