package middleware

import (
	"context"
	"hermes-relay/internal/cqrs"
	"log/slog"
)

func WithLogging(level slog.Level) cqrs.CommandRouter {
	return func(ctx context.Context, action *cqrs.Message, publisher cqrs.PublishFunc) (*cqrs.Message, error) {
		slog.Log(context.Background(), level, "message received", "message", action)
		return nil, nil
	}
}
