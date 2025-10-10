package main

import (
	"hermes-relay/internal/domain/commands"
	"hermes-relay/internal/domain/events"
	"hermes-relay/internal/handlers"
	"hermes-relay/internal/handlers/middleware"
	"hermes-relay/internal/utils/dispatch"
	"log"
	"log/slog"
	"net/http"
	"os"
)

var router = dispatch.MakeCombinedRouter(
	middleware.WithLogging(slog.LevelDebug),
	commands.Router,
	events.Router,
)

func main() {
	var publisher = dispatch.NewInMemoryPublisher()
	publisher.Subscribe(router)

	// Todo: Change to env var
	setupLogger(slog.LevelDebug)

	http.HandleFunc("/command", handlers.CommandHandler(publisher))
	http.HandleFunc("/event", handlers.EventHandler(publisher))

	http.HandleFunc("/ws", handlers.WebSocketHandler(publisher))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupLogger(level slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug, // Add file:line in debug mode
	}))

	slog.SetDefault(logger)
}
