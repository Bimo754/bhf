package api

import (
	"log"
	"net/http"

	"github.com/diamond/bhf/internal/engine"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow development and dashboard access
	},
}

// WSHandler manages WebSocket connections for real-time terminal output streaming.
type WSHandler struct {
	hub *engine.Hub
}

// NewWSHandler creates a new WSHandler.
func NewWSHandler(hub *engine.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// StreamLogs handles WebSocket upgrade and subscribes the client to live scan logs.
func (h *WSHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	scanID := chi.URLParam(r, "id")
	if scanID == "" {
		http.Error(w, "Scan ID required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	h.hub.Subscribe(scanID, conn)
	defer h.hub.Unsubscribe(scanID, conn)

	// Keep alive until client closes
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
