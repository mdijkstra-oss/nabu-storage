package handlers

import (
	"context"
	"github.com/gorilla/websocket"
	commands2 "hermes-relay/internal/commands"
	"hermes-relay/internal/utils"
	"net/http"
	"time"
)

type ErrorResponse struct {
	Error []string `json:"error"`
	Type  string   `json:"type"`
}

func WebSocketHandler(publisher *commands2.InMemoryPublisher) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Configure this properly in production
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		SetupWebSocketForwarder(publisher, conn)

		// Keep connection alive
		select {}
	}
}

func SetupWebSocketForwarder(publisher *commands2.InMemoryPublisher, conn *websocket.Conn) {
	// Subscribe to outgoing events
	publisher.Subscribe(func(ctx context.Context, event *commands2.Message, pub commands2.PublishFunc) (*commands2.Message, error) {
		if event.Type == commands2.DomainEvent || event.Type == commands2.SystemEvent {
			utils.WarnErr(conn.WriteJSON(event))
		}

		return nil, nil
	})

	// Handle incoming messages from WebSocket
	go func() {
		for {
			var msg commands2.Message
			decodeErr := conn.ReadJSON(&msg)
			if decodeErr != nil {
				continue // Connection closed
			}

			msg.Type = commands2.Command
			msg.Timestamp = time.Now()

			_, publishErr := publisher.Publish(context.Background(), &msg)
			if publishErr != nil {
				utils.WarnErr(conn.WriteJSON(ErrorResponse{
					Error: []string{publishErr.Error()},
					Type:  utils.GetErrorType(publishErr),
				}))
				continue
			}

			// Result already sent by subscriber above
		}
	}()
}
