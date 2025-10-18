package middleware

import (
	"context"
	"hermes-relay/internal/commands"
	"log/slog"
)

func WithLogging(level slog.Level) commands.CommandRouter {
	return func(ctx context.Context, action *commands.Message, publisher commands.PublishFunc) (*commands.Message, error) {
		slog.Log(context.Background(), level, "message received", "message", action)
		return nil, nil
	}
}
