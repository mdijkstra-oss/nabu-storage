package http

import (
	"github.com/gorilla/websocket"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/lib/utils"
	"net/http"
)

func WebSocketHandler(publish dispatch.PublishFunc, subscribe func(dispatch.CommandRouter)) http.HandlerFunc {
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

func handleWebSocket(conn *websocket.Conn, publish dispatch.PublishFunc, subscribe func(dispatch.CommandRouter)) {
	// Forward domain/system events to client
	subscribe(dispatch.LimitOnType(
		commands.DomainEvent,
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

func forwardToWebSocket(conn *websocket.Conn) dispatch.CommandRouter {
	return func(msg *commands.AnyMessage, pub dispatch.PublishFunc) (*commands.AnyMessage, error) {
		utils.WarnErr(conn.WriteJSON(msg))
		return nil, nil
	}
}
