package main

import (
	"context"
	"hermes-relay/internal/commands"
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

var router = commands.MakeCombinedRouter(
	middleware.WithLogging(slog.LevelDebug),
	pingpong.Router,
	TempFileCommand,
)

func TempFileCommand(ctx context.Context, message *commands.Message, _ commands.PublishFunc) (*commands.Message, error) {
	if (message.Action == "PatchFile") && message.Payload != nil {
		return commands.CommandToDomainEvent(message), nil
	}

	return nil, nil
}

func main() {

	// Todo: Change to env var
	setupLogger(slog.LevelDebug)

	var store = persistence.NewStore()

	existingEvents := utils.Must(events.LoadEvents("files/events.json"))

	utils.MustNotError(store.ApplyEvents(existingEvents))

	var publisher = commands.NewInMemoryPublisher()
	publisher.Subscribe(router)

	// In some future, we can add DomainEvent routing too, to generate new actions eg FileUploaded, TranscodeFile, TranscodedFile etc
	publisher.Subscribe(commands.LimitOnType(commands.DomainEvent, func(ctx context.Context, message *commands.Message, publisher commands.PublishFunc) (*commands.Message, error) {
		err := store.ApplyEvent(*message)
		// Todo: write event
		return nil, err
	}))

	// Todo: in some future, for both auth etc
	// Events would be internal I think, but still, validate (else it can write any entity now)
	http.HandleFunc("/commands", handlers.CommandHandler(publisher))
	http.HandleFunc("/events", handlers.EventHandler(publisher))

	http.HandleFunc("/ws", handlers.WebSocketHandler(publisher))

	http.HandleFunc("/queries/files/", handlers.RESTHandler[file.File](store))

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupLogger(level slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
}
