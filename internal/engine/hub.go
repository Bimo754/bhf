package engine

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub manages WebSocket client connections subscribed to specific scan execution streams.
type Hub struct {
	mu          sync.RWMutex
	subscribers map[string]map[*websocket.Conn]bool
}

// NewHub initializes an empty Hub.
func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[string]map[*websocket.Conn]bool),
	}
}

// Subscribe adds a websocket connection to a scan's broadcast channel.
func (h *Hub) Subscribe(scanID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.subscribers[scanID]; !ok {
		h.subscribers[scanID] = make(map[*websocket.Conn]bool)
	}
	h.subscribers[scanID][conn] = true
}

// Unsubscribe removes a websocket connection.
func (h *Hub) Unsubscribe(scanID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conns, ok := h.subscribers[scanID]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.subscribers, scanID)
		}
	}
}

// Broadcast sends a stream message to all active clients watching a scan.
func (h *Hub) Broadcast(msg StreamMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	conns, ok := h.subscribers[msg.ScanID]
	if !ok {
		return
	}

	for conn := range conns {
		_ = conn.WriteJSON(msg)
	}
}
