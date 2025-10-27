package http

import (
	"github.com/gorilla/websocket"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/lib/utils"
	"net/http"
)

func WebSocketHandler(publish cqrs.PublishFunc, subscribe func(cqrs.CommandRouter)) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Todo: Configure properly in production
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		handleWebSocket(conn, publish, subscribe)
	}
}

func handleWebSocket(conn *websocket.Conn, publish cqrs.PublishFunc, subscribe func(cqrs.CommandRouter)) {
	// Forward domain/system events to client
	subscribe(cqrs.LimitOnType(
		cqrs.DomainEvent,
		forwardToWebSocket(conn),
	))

	// Handle incoming commands from client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return // Connection closed
		}

		// Commands do not need ctx, if something is async it will time out on its own
		// Only fast acting things mostly, and video encoding etc would still need to continue even if user loses connection
		response := ProcessCommand(Request{Body: message}, publish)

		// Send response (errors or results)
		if len(response.Body) > 0 {
			utils.WarnErr(conn.WriteMessage(websocket.TextMessage, response.Body))
		}
	}
}

func forwardToWebSocket(conn *websocket.Conn) cqrs.CommandRouter {
	return func(msg *cqrs.AnyMessage, pub cqrs.PublishFunc) (*cqrs.AnyMessage, error) {
		utils.WarnErr(conn.WriteJSON(msg))
		return nil, nil
	}
}
