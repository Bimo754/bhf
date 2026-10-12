package models

import "time"

// Asset represents a discovered subdomain, web endpoint, or service.
type Asset struct {
	ID           string    `json:"id"`
	TargetID     string    `json:"target_id"`
	ScanID       string    `json:"scan_id"`
	Host         string    `json:"host"`
	URL          string    `json:"url,omitempty"`
	IP           string    `json:"ip,omitempty"`
	Port         int       `json:"port,omitempty"`
	StatusCode   int       `json:"status_code,omitempty"`
	Title        string    `json:"title,omitempty"`
	Technologies []string  `json:"technologies,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
