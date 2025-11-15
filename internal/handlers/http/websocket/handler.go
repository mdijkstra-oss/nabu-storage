package websocket

import (
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/patches"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	"net/http"
)

type Message struct {
	Type      string `json:"type"`
	ProjectID string `json:"projectId"`
	Data      any    `json:"data"`
}

func Handler(hub *Hub, registryState *registry.RegistryState, subscribe func(dispatch.CommandRouter) func()) http.HandlerFunc {
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

		handleConnection(conn, projectID, hub, registryState, subscribe)
	}
}

func handleConnection(
	conn *websocket.Conn,
	projectID string,
	hub *Hub,
	registryState *registry.RegistryState,
	subscribe func(dispatch.CommandRouter) func(),
) {
	hub.Register(projectID, conn)
	defer hub.Unregister(projectID, conn)

	sendInitialSnapshot(conn, projectID, registryState)

	unsubscribe := subscribe(dispatch.LimitOnType(
		commands.SystemEvent,
		dispatch.CombineRouters(
			forwardProjectEvent[patches.PatchEventPayload](conn, projectID, patches.ProjectPatched, "patch"),
			forwardProjectEvent[patches.SnapshotEventPayload](conn, projectID, patches.ProjectSnapshot, "snapshot"),
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

func sendInitialSnapshot(conn *websocket.Conn, projectID string, registryState *registry.RegistryState) {
	utils.GuardWith(func() {
		project := registryState.GetProject(projectID)
		if project == nil {
			slog.Warn("project not found for initial snapshot", "projectID", projectID)
			return
		}

		msg := Message{
			Type:      "snapshot",
			ProjectID: projectID,
			Data:      project,
		}

		utils.ShouldWork(conn.WriteJSON(msg))
	}, "projectID", projectID, "operation", "sendInitialSnapshot")
}

func forwardProjectEvent[T patches.ProjectEventPayload](
	conn *websocket.Conn,
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

			utils.ShouldWork(conn.WriteJSON(message))
		}, "projectID", projectID, "action", action, "operation", "forwardProjectEvent")

		return nil, nil
	})
}
