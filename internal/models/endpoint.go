package models

import "time"

// Endpoint represents a discovered web path or URL on a host.
type Endpoint struct {
	ID            string    `json:"id"`
	TargetID      string    `json:"target_id"`
	ScanID        string    `json:"scan_id,omitempty"`
	Host          string    `json:"host"`
	URL           string    `json:"url"`
	Path          string    `json:"path"`
	Method        string    `json:"method,omitempty"`
	StatusCode    int       `json:"status_code,omitempty"`
	ContentLength int       `json:"content_length,omitempty"`
	Title         string    `json:"title,omitempty"`
	SourceTool    string    `json:"source_tool,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
