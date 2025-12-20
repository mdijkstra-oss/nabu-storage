package websocket

import (
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/cqrs/dispatch"
	"hermes-relay/internal/cqrs/patches"
	"hermes-relay/internal/domain/entities/file"
	"hermes-relay/internal/domain/entities/project"
	"hermes-relay/internal/domain/projections/registry"
	"hermes-relay/internal/lib/utils"
	"log/slog"
	"net/http"
	"time"
)

const (
	pongWait   = 30 * time.Second
	pingPeriod = (pongWait * 9) / 10
	writeWait  = 10 * time.Second
)

type Message struct {
	Type      string `json:"type"`
	ProjectID string `json:"project_id"`
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

		fileID := r.URL.Query().Get("fileId")

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { utils.ShouldWork(conn.Close()) }()

		handleConnection(conn, projectID, fileID, hub, registryState, subscribe)
	}
}

func setupPongHandler(conn *websocket.Conn) {
	utils.ShouldWork(conn.SetReadDeadline(time.Now().Add(pongWait)))
	conn.SetPongHandler(func(string) error {
		utils.ShouldWork(conn.SetReadDeadline(time.Now().Add(pongWait)))
		return nil
	})
}

func startPingSender(conn *websocket.Conn, done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			utils.ShouldWork(conn.SetWriteDeadline(time.Now().Add(writeWait)))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
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
	fileID string,
	hub *Hub,
	registryState *registry.RegistryState,
	subscribe func(dispatch.CommandRouter) func(),
) {
	hub.Register(projectID, conn)
	defer hub.Unregister(projectID, conn)

	setupPongHandler(conn)

	done := make(chan struct{})
	defer close(done)
	go startPingSender(conn, done)

	sendInitialSnapshot(conn, projectID, fileID, registryState)

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

func sendInitialSnapshot(conn *websocket.Conn, projectID string, fileID string, registryState *registry.RegistryState) {
	utils.GuardWith(func() {
		project := registryState.GetProject(projectID)
		if project == nil {
			slog.Warn("project not found for initial snapshot", "projectID", projectID)
			return
		}

		filteredProject := filterProjectForFile(project, fileID)

		msg := Message{
			Type:      "snapshot",
			ProjectID: projectID,
			Data:      filteredProject,
		}

		utils.ShouldWork(conn.WriteJSON(msg))
	}, "projectID", projectID, "operation", "sendInitialSnapshot")
}

func filterProjectForFile(p *project.Project, fileID string) *project.Project {
	if fileID == "" {
		return p
	}

	filteredFiles := make(map[string]file.File)
	for id, f := range p.Files {
		if id == fileID {
			filteredFiles[id] = withCodeCount(f)
		} else {
			filteredFiles[id] = withCodeCountOnly(f)
		}
	}

	filtered := *p
	filtered.Files = filteredFiles
	return &filtered
}

func withCodeCount(f file.File) file.File {
	f.CodeCount = len(f.Codes)
	return f
}

func withCodeCountOnly(f file.File) file.File {
	f.CodeCount = len(f.Codes)
	f.Content = ""
	f.Codes = nil
	return f
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
