package websocket

import (
	"github.com/gorilla/websocket"
	"hermes-relay/internal/lib/utils"
	"sync"
)

// Todo: Add tests for Hub
// - Concurrent Register/Unregister
// - Broadcast to multiple connections
// - Connection cleanup on write errors
// - IsActive with multiple projects

type Hub struct {
	connections map[string]map[*websocket.Conn]bool
	mu          sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]map[*websocket.Conn]bool),
	}
}

func (h *Hub) Register(projectID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections[projectID] == nil {
		h.connections[projectID] = make(map[*websocket.Conn]bool)
	}
	h.connections[projectID][conn] = true
}

func (h *Hub) Unregister(projectID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections[projectID] != nil {
		delete(h.connections[projectID], conn)
		if len(h.connections[projectID]) == 0 {
			delete(h.connections, projectID)
		}
	}
}

func (h *Hub) IsActive(projectID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections[projectID]) > 0
}

func (h *Hub) Broadcast(projectID string, messageType int, data []byte) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.connections[projectID]))
	for conn := range h.connections[projectID] {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()

	for _, conn := range conns {
		utils.GuardWith(func() {
			if err := conn.WriteMessage(messageType, data); err != nil {
				h.Unregister(projectID, conn)
				conn.Close()
			}
		}, "projectID", projectID, "operation", "broadcast")
	}
}
