package main

import (
	"context"
	"hermes-relay/internal/commands"
	"hermes-relay/internal/domain/entities/code"
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
	commands.LimitOnType(commands.Command, CommandToEvent),
)

func CommandToEvent(ctx context.Context, message *commands.Message, _ commands.PublishFunc) (*commands.Message, error) {
	_, _, parseErr := commands.ParseAction(message.Action)
	if parseErr != nil {
		return nil, parseErr
	}

	return commands.CommandToDomainEvent(message), nil
}

func main() {

	// Todo: Change to env var
	setupLogger(slog.LevelDebug)

	//existingEvents := utils.Must(events.LoadEvents("files/events.json"))
	existingEvents := []commands.Message{
		*utils.Must(file.CreateFileEventFromPath("files/rutte-lang.md")),
	}

	utils.MustNotError(persistence.StoreForNoun("File").ApplyEvents(existingEvents))

	var publisher = commands.NewInMemoryPublisher()
	publisher.Subscribe(router)

	// In some future, we can add DomainEvent routing too, to generate new actions eg FileUploaded, TranscodeFile, TranscodedFile etc
	publisher.Subscribe(commands.LimitOnType(commands.DomainEvent, func(ctx context.Context, message *commands.Message, publisher commands.PublishFunc) (*commands.Message, error) {

		verb, noun, parseErr := commands.ParseAction(message.Action)
		if parseErr != nil {
			return nil, parseErr
		}

		store := persistence.StoreForNoun(noun)
		if verb == "DELETE" {
			store.DeleteByID(message.ID)
			return nil, nil
		}

		err := store.ApplyEvent(*message)

		if noun == "File" && verb == "Patched" {
			// should not do this ofc. could be in other files tooo....
			//cleanupOrphanedCodes(message.AggregateID)
		}

		// Todo: write event
		return nil, err
	}))

	// Todo: in some future, for both auth etc
	// Events would be internal I think, but still, validate (else it can write any entity now)
	// todo: ensure end slashes are optional
	http.HandleFunc("/commands", handlers.CommandHandler(publisher))
	http.HandleFunc("/events", handlers.EventHandler(publisher))

	http.HandleFunc("/ws/", handlers.WebSocketHandler(publisher))

	http.HandleFunc("/queries/files/", handlers.RESTHandler[file.File](persistence.StoreForNoun(utils.GetStructName(file.File{}))))
	http.HandleFunc("/queries/codes/", handlers.RESTHandler[code.Code](persistence.StoreForNoun(utils.GetStructName(code.Code{}))))

	log.Fatal(http.ListenAndServe(":8080", CORS(http.DefaultServeMux)))
}

func setupLogger(level slog.Level) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	slog.SetDefault(logger)
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func cleanupOrphanedCodes(fileID string) {
	fileStore := persistence.StoreForNoun("File")
	codeStore := persistence.StoreForNoun("Code")

	// Get updated file
	updatedFile, err := persistence.GetByID[file.File](fileStore, fileID)
	if err != nil {
		slog.Error("Failed to get file", "fileID", fileID, "error", err)
		return
	}

	// Get all code IDs from the file
	fileCodeIDs := make(map[string]bool)
	for _, c := range updatedFile.Attributes.Codes {
		fileCodeIDs[c.ID] = true
	}

	// Get all codes from the code store
	allCodes, err := persistence.GetAll[code.Code](codeStore)
	if err != nil {
		slog.Error("Failed to get all codes", "error", err)
		return
	}

	// Delete codes that are no longer in the file
	for _, c := range allCodes {
		if !fileCodeIDs[c.ID] {
			codeStore.DeleteByID(c.ID)
			slog.Info("Deleted orphaned code", "codeID", c.ID, "fileID", fileID)
		}
	}
}
