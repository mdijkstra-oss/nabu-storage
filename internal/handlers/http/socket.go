package http

import (
	"context"
	"github.com/gorilla/websocket"
	"hermes-relay/internal/cqrs"
	"hermes-relay/internal/lib/utils"
	"net/http"
	"time"
)

func WebSocketHandler(publish cqrs.PublishFunc, subscribe func(cqrs.CommandRouter)) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Configure properly in production
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Forward domain/system events to client
	// Todo: Ofc not all on multiple clients etc
	subscribe(cqrs.LimitOnType(
		cqrs.DomainEvent,
		forwardToWebSocket(conn),
	))

	// Handle incoming commands from client
	for {
		var msg cqrs.AnyMessage
		if err := conn.ReadJSON(&msg); err != nil {
			return // Connection closed
		}

		msg.Type = cqrs.Command
		msg.Timestamp = time.Now()

		result, err := publish(ctx, &msg)
		if err != nil {
			sendError(conn, err)
			continue
		}

		if result != nil {
			utils.WarnErr(conn.WriteJSON(result))
		}
	}
}

func forwardToWebSocket(conn *websocket.Conn) cqrs.CommandRouter {
	return func(ctx context.Context, msg *cqrs.AnyMessage, pub cqrs.PublishFunc) (*cqrs.AnyMessage, error) {
		utils.WarnErr(conn.WriteJSON(msg))
		return nil, nil
	}
}

func sendError(conn *websocket.Conn, err error) {
	response := ErrorResponse{Message: err.Error()}

	if ve, ok := err.(*utils.ValidationError); ok {
		response.Fields = ve.Fields
	}

	utils.WarnErr(conn.WriteJSON(response))
}
