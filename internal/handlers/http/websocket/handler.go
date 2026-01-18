package websocket

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/patches"
	"hermes-relay/internal/cqrs/projection"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/lib/utils"
)

const (
	pongWait   = 30 * time.Second
	pingPeriod = (pongWait * 9) / 10
	writeWait  = 10 * time.Second
)

type connWriter struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *connWriter) writeJSON(v any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	utils.ShouldWork(w.conn.SetWriteDeadline(time.Now().Add(writeWait)))
	return w.conn.WriteJSON(v)
}

func (w *connWriter) writePing() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	utils.ShouldWork(w.conn.SetWriteDeadline(time.Now().Add(writeWait)))
	return w.conn.WriteMessage(websocket.PingMessage, nil)
}

type Message struct {
	Type      string `json:"type"`
	ProjectID string `json:"project_id"`
	Data      any    `json:"data"`
}

func Handler(hub *Hub, store *registry.Store, subscribe func(dispatch.CommandRouter) func()) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		if projectID == "" {
			http.Error(w, "projectId required", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { utils.ShouldWork(conn.Close()) }()

		handleConnection(conn, projectID, hub, store, subscribe)
	}
}

func setupPongHandler(conn *websocket.Conn) {
	utils.ShouldWork(conn.SetReadDeadline(time.Now().Add(pongWait)))
	conn.SetPongHandler(func(string) error {
		utils.ShouldWork(conn.SetReadDeadline(time.Now().Add(pongWait)))
		return nil
	})
}

func startPingSender(writer *connWriter, done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := writer.writePing(); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

func handleConnection(
	conn *websocket.Conn,
	projectID string,
	hub *Hub,
	store *registry.Store,
	subscribe func(dispatch.CommandRouter) func(),
) {
	hub.Register(projectID, conn)
	defer hub.Unregister(projectID, conn)

	writer := &connWriter{conn: conn}

	setupPongHandler(conn)

	done := make(chan struct{})
	defer close(done)
	go startPingSender(writer, done)

	sendInitialSnapshot(writer, projectID, store)

	unsubscribe := subscribe(dispatch.LimitOnType(
		commands.SystemEvent,
		dispatch.CombineRouters(
			forwardProjectEvent[patches.PatchEventPayload](writer, projectID, patches.ProjectPatched, "patch"),
			forwardProjectEvent[patches.SnapshotEventPayload](writer, projectID, patches.ProjectSnapshot, "snapshot"),
		),
	))
	defer unsubscribe()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				slog.Warn("websocket connection error", "projectID", projectID, "error", err)
			}
			return
		}
	}
}

func sendInitialSnapshot(writer *connWriter, projectID string, store *registry.Store) {
	utils.GuardWith(func() {
		proj := projection.Read(store, func(r *registry.Registry) *project.Project {
			return registry.GetProject(r, projectID)
		})
		if proj == nil {
			slog.Warn("project not found for initial snapshot", "projectID", projectID)
			return
		}

		msg := Message{
			Type:      "snapshot",
			ProjectID: projectID,
			Data:      proj,
		}

		utils.ShouldWork(writer.writeJSON(msg))
	}, "projectID", projectID, "operation", "sendInitialSnapshot")
}

func forwardProjectEvent[T patches.ProjectEventPayload](
	writer *connWriter,
	projectID string,
	action commands.Action,
	messageType string,
) dispatch.CommandRouter {
	return dispatch.LimitOnAction(action, func(msg *commands.AnyMessage, _ dispatch.PublishFunc) (*commands.AnyMessage, error) {
		utils.GuardWith(func() {
			var payload T
			if err := commands.UnmarshallPayload(msg, &payload); err != nil {
				slog.Error("failed to unmarshal payload", "action", action, "error", err)
				return
			}

			if payload.GetProjectID() != projectID {
				return
			}

			message := Message{
				Type:      messageType,
				ProjectID: payload.GetProjectID(),
				Data:      payload.GetData(),
			}

			utils.ShouldWork(writer.writeJSON(message))
		}, "projectID", projectID, "action", action, "operation", "forwardProjectEvent")

		return nil, nil
	})
}
