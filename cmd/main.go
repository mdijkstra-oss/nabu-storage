package main

import (
	"context"
	commands2 "hermes-relay/internal/commands"
	"hermes-relay/internal/commands/events"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/pingpong"
	"hermes-relay/internal/handlers"
	"hermes-relay/internal/handlers/middleware"
	"hermes-relay/internal/persistence"
	"hermes-relay/internal/utils"
	"log"
	"log/slog"
	"net/http"
	"os"
)

var router = commands2.MakeCombinedRouter(
	middleware.WithLogging(slog.LevelDebug),
	pingpong.Router,
)

func main() {

	// Todo: Change to env var
	setupLogger(slog.LevelDebug)

	var store = persistence.NewStore()

	existingEvents := utils.Must(events.LoadEvents("files/events.json"))

	utils.MustNotError(store.ApplyEvents(existingEvents))

	var publisher = commands2.NewInMemoryPublisher()
	publisher.Subscribe(router)

	// In some future, we can add DomainEvent routing too, to generate new actions eg FileUploaded, TranscodeFile, TranscodedFile etc
	publisher.Subscribe(commands2.LimitOnType(commands2.DomainEvent, func(ctx context.Context, message *commands2.Message, publisher commands2.PublishFunc) (*commands2.Message, error) {
		err := store.ApplyEvent(*message)
		// Todo: write event
		return nil, err
	}))

	// Todo: in some future, for both auth etc
	// Events would be internal I think, but still, validate (else it can write any entity now)
	http.HandleFunc("/command", handlers.CommandHandler(publisher))
	http.HandleFunc("/event", handlers.EventHandler(publisher))

	http.HandleFunc("/ws", handlers.WebSocketHandler(publisher))

	http.HandleFunc("/queries/files/", handlers.RESTHandler[file.File](store))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupLogger(level slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug, // Add file:line in debug mode
	}))

	slog.SetDefault(logger)
}
