package websocket

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	connections map[string]map[*websocket.Conn]bool
	mu          sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[string]map[*websocket.Conn]bool),
	}
}

func Register(h *Hub, projectID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections[projectID] == nil {
		h.connections[projectID] = make(map[*websocket.Conn]bool)
	}
	h.connections[projectID][conn] = true
}

func Unregister(h *Hub, projectID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.connections[projectID] != nil {
		delete(h.connections[projectID], conn)
		if len(h.connections[projectID]) == 0 {
			delete(h.connections, projectID)
		}
	}
}
