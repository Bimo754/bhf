package engine

import (
	"time"

	"github.com/diamond/bhf/internal/models"
)

// StreamMessage represents a real-time event sent over WebSockets.
type StreamMessage struct {
	Type      string           `json:"type"` // "log", "status", "complete", "error"
	ScanID    string           `json:"scan_id"`
	Status    models.ScanStatus `json:"status,omitempty"`
	Stream    models.LogStream  `json:"stream,omitempty"`
	Line      string           `json:"line,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
}
