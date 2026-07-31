package websocket

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"nabu-storage/internal/domain"
	"nabu-storage/internal/domain/files"
	"nabu-storage/internal/lib/utils"
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

func writeJSON(w *connWriter, v any) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	utils.ShouldWork(w.conn.SetWriteDeadline(time.Now().Add(writeWait)))

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return w.conn.WriteMessage(websocket.TextMessage, data)
}

func writePing(w *connWriter) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	utils.ShouldWork(w.conn.SetWriteDeadline(time.Now().Add(writeWait)))
	return w.conn.WriteMessage(websocket.PingMessage, nil)
}

func allowAllOrigins(*http.Request) bool {
	return true
}

func Handler(baseDir string) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: allowAllOrigins,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		projectID := chi.URLParam(r, "projectId")
		if !utils.ValidID(projectID) {
			http.Error(w, "invalid projectId", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { utils.ShouldWork(conn.Close()) }()

		handleConnection(conn, projectID, baseDir)
	}
}

func handleConnection(conn *websocket.Conn, projectID, baseDir string) {
	writer := &connWriter{conn: conn}

	setupPongHandler(conn)

	done := make(chan struct{})
	defer close(done)
	go startPingSender(writer, done)

	sendInitialFiles(writer, projectID, baseDir)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if isUnexpectedClose(err) {
				slog.Warn("websocket connection error", "projectID", projectID, "error", err)
			}
			return
		}
	}
}

func isUnexpectedClose(err error) bool {
	return !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)
}

func setupPongHandler(conn *websocket.Conn) {
	utils.ShouldWork(conn.SetReadDeadline(time.Now().Add(pongWait)))
	conn.SetPongHandler(createPongHandler(conn))
}

func createPongHandler(conn *websocket.Conn) func(string) error {
	return func(string) error {
		utils.ShouldWork(conn.SetReadDeadline(time.Now().Add(pongWait)))
		return nil
	}
}

func startPingSender(writer *connWriter, done chan struct{}) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := writePing(writer); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

func sendInitialFiles(writer *connWriter, projectID, baseDir string) {
	utils.GuardWith(
		sendInitialFilesWork(writer, projectID, baseDir),
		"projectID", projectID, "operation", "sendInitialFiles",
	)
}

func sendInitialFilesWork(writer *connWriter, projectID, baseDir string) func() {
	return func() {
		files.SeedRequiredFiles(baseDir, projectID)

		fileNames, err := files.List(baseDir, projectID)
		if err != nil {
			slog.Error("failed to list files", "projectID", projectID, "error", err)
			return
		}

		sorted := files.SortForInitialSend(fileNames)

		meta := domain.Command{Action: domain.SyncMeta, FileCount: len(sorted)}
		if err := writeJSON(writer, meta); err != nil {
			return
		}

		for _, name := range sorted {
			if err := sendFileAsCreate(writer, projectID, baseDir, name); err != nil {
				return
			}
		}
	}
}

func sendFileAsCreate(writer *connWriter, projectID, baseDir, name string) error {
	content, err := files.Read(baseDir, projectID, name)
	if err != nil {
		slog.Warn("failed to read file for initial send", "projectID", projectID, "file", name, "error", err)
		return nil
	}

	cmd := domain.Command{
		Action:  domain.WriteFile,
		Path:    name,
		Content: content,
	}

	if err := writeJSON(writer, cmd); err != nil {
		slog.Warn("failed to send initial file", "projectID", projectID, "file", name, "error", err)
		return err
	}
	return nil
}
