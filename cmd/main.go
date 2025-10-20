package main

import (
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/handlers"
	"hermes-relay/internal/handlers/middleware"
	"hermes-relay/internal/projection"
	"hermes-relay/internal/utils"
	"log"
	"log/slog"
	"net/http"
	"os"
)

var commandRouter = cqrs.CombineRouters(
	// Entity-specific command handlers
	code.Router,
	file.Router,
)

func main() {

	// Todo: Change to env var
	setupLogger(slog.LevelDebug)

	//existingEvents := utils.Must(events.LoadEvents("files/events.json"))
	existingEvents := []cqrs.Message{
		*utils.Must(file.CreateFileEventFromPath("files/rutte-lang.md")),
	}

	utils.MustNotError(projection.FileStore.ApplyEvents(existingEvents))

	var publisher = cqrs.NewInMemoryPublisher()
	publisher.Subscribe(commandRouter)

	// Domain event handler - applies events to stores using reducers
	publisher.Subscribe(cqrs.LimitOnType(cqrs.DomainEvent, cqrs.ReadOnlyRoute(projection.ApplyEventToStore)))

	publisher.Subscribe(middleware.WithLogging(slog.LevelDebug))

	// Todo: in some future, for both auth etc
	// Events would be internal I think, but still, validate (else it can write any entity now)
	// todo: ensure end slashes are optional
	http.HandleFunc("/commands", handlers.CommandHandler(publisher))
	http.HandleFunc("/events", handlers.EventHandler(publisher))

	http.HandleFunc("/ws/", handlers.WebSocketHandler(publisher))

	http.HandleFunc("/queries/files/", handlers.RESTHandler[file.File](projection.FileStore))
	http.HandleFunc("/queries/codes/", handlers.RESTHandler[code.Code](projection.CodeStore))

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
