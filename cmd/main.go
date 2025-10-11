package main

import (
	"context"
	"hermes-relay/internal/domain/commands"
	"hermes-relay/internal/domain/entities/file"
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
	var store = dispatch.NewStore()

	existingEvents := utils.Must(dispatch.LoadEvents("files/events.json"))

	utils.MustNotError(store.ApplyEvents(existingEvents))

	var publisher = dispatch.NewInMemoryPublisher()
	publisher.Subscribe(router)

	// In some future, we can add Event routing too, to generate new actions eg FileUploaded, TranscodeFile, TranscodedFile etc
	publisher.Subscribe(dispatch.LimitOnType(dispatch.Event, func(ctx context.Context, message *dispatch.Message, publisher dispatch.PublishFunc) (*dispatch.Message, error) {
		err := store.ApplyEvent(*message)
		return nil, err
	}))

	var firstFile = file.File{}
	utils.MustNotError(store.GetByID("file-001", &firstFile))
	log.Println(firstFile)

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
