package main

import (
	"context"
	"hermes-relay/internal/domain/commands"
	"hermes-relay/internal/handlers"
	"hermes-relay/internal/handlers/middleware"
	"hermes-relay/internal/utils"
	"hermes-relay/internal/utils/dispatch"
	"log"
	"log/slog"
	"net/http"
	"os"
)

var router = dispatch.MakeCombinedRouter(
	middleware.WithLogging(slog.LevelDebug),
	commands.Router,
)

func main() {

	// Todo: Change to env var
	setupLogger(slog.LevelDebug)

	var store = dispatch.NewStore()

	existingEvents := utils.Must(dispatch.LoadEvents("files/events.json"))

	utils.MustNotError(store.ApplyEvents(existingEvents))

	var publisher = dispatch.NewInMemoryPublisher()
	publisher.Subscribe(router)

	// In some future, we can add DomainEvent routing too, to generate new actions eg FileUploaded, TranscodeFile, TranscodedFile etc
	publisher.Subscribe(dispatch.LimitOnType(dispatch.DomainEvent, func(ctx context.Context, message *dispatch.Message, publisher dispatch.PublishFunc) (*dispatch.Message, error) {
		err := store.ApplyEvent(*message)
		// Todo: write event
		return nil, err
	}))

	// Todo: in some future, for both auth etc
	// Events would be internal I think, but still, validate (else it can write any entity now)
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
