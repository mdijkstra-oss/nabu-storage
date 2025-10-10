package middleware

import (
	"context"
	"hermes-relay/internal/utils/dispatch"
	"log/slog"
)

func WithLogging(level slog.Level) dispatch.CommandRouter {
	return func(ctx context.Context, action *dispatch.Message, publisher dispatch.PublishFunc) (*dispatch.Message, error) {
		slog.Log(context.Background(), level, "message received", "message", action)
		return nil, nil
	}
}
