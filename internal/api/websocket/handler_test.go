package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	gws "github.com/gorilla/websocket"
	"nabu-storage/internal/domain"
	"nabu-storage/internal/domain/files"
	th "nabu-storage/internal/lib/testutil"
)

func TestInitialSync(t *testing.T) {
	const projectID = "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		Name        string
		Files       map[string]string
		Unreadable  string
		ExpectPaths []string
	}{
		{
			Name:        "unknown project sends an explicit zero count",
			ExpectPaths: []string{},
		},
		{
			Name: "sends every file in alphabetical order",
			Files: map[string]string{
				"notes.md":   "# Notes",
				"a.md":       "first",
				"journal.md": "# Journal",
			},
			ExpectPaths: []string{"a.md", "journal.md", "notes.md", "preferences.md", "settings.hidden.md"},
		},
		{
			Name: "count excludes files that cannot be read",
			Files: map[string]string{
				"readable.md": "ok",
				"locked.md":   "secret",
			},
			Unreadable:  "locked.md",
			ExpectPaths: []string{"preferences.md", "readable.md", "settings.hidden.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if tt.Unreadable != "" && os.Geteuid() == 0 {
				t.Skip("root bypasses file permissions")
			}

			baseDir := t.TempDir()
			for path, content := range tt.Files {
				th.AssertError(t, files.Write(baseDir, projectID, path, content), "", "setup write")
			}
			if tt.Unreadable != "" {
				path := files.FilePath(baseDir, projectID, tt.Unreadable)
				th.AssertError(t, os.Chmod(path, 0000), "", "setup chmod")
			}

			conn, cleanup := dialProject(t, baseDir, projectID)
			defer cleanup()

			meta := readFrame(t, conn)
			th.AssertEqual(t, meta["action"], string(domain.SyncMeta), "meta action")
			count, ok := meta["fileCount"]
			if !ok {
				t.Fatal("fileCount missing from SyncMeta frame")
			}
			th.AssertEqual(t, count, float64(len(tt.ExpectPaths)), "file count")

			paths := make([]string, 0, len(tt.ExpectPaths))
			for range tt.ExpectPaths {
				frame := readFrame(t, conn)
				th.AssertEqual(t, frame["action"], string(domain.WriteFile), "file action")
				paths = append(paths, frame["path"].(string))
			}
			th.AssertEqual(t, paths, tt.ExpectPaths, "sent paths")

			assertNoMoreFrames(t, conn)
		})
	}
}

func dialProject(t *testing.T, baseDir, projectID string) (*gws.Conn, func()) {
	t.Helper()

	router := chi.NewRouter()
	router.Get("/ws/{projectId}", Handler(baseDir))
	srv := httptest.NewServer(router)

	conn, resp, err := gws.DefaultDialer.Dial(strings.Replace(srv.URL, "http://", "ws://", 1)+"/ws/"+projectID, nil)
	if err != nil {
		srv.Close()
		t.Fatalf("dial: %v", err)
	}
	th.AssertEqual(t, resp.StatusCode, http.StatusSwitchingProtocols, "upgrade status")

	return conn, func() {
		th.AssertError(t, conn.Close(), "", "close conn")
		srv.Close()
	}
}

func readFrame(t *testing.T, conn *gws.Conn) map[string]any {
	t.Helper()

	th.AssertError(t, conn.SetReadDeadline(time.Now().Add(5*time.Second)), "", "read deadline")
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}

	var frame map[string]any
	if err := json.Unmarshal(data, &frame); err != nil {
		t.Fatalf("decode frame %q: %v", data, err)
	}
	return frame
}

func assertNoMoreFrames(t *testing.T, conn *gws.Conn) {
	t.Helper()

	th.AssertError(t, conn.SetReadDeadline(time.Now().Add(200*time.Millisecond)), "", "read deadline")
	if _, data, err := conn.ReadMessage(); err == nil {
		t.Fatalf("expected no further frames, got %q", data)
	}
}
