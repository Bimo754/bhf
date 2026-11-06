package models

import "time"

// TrafficRecord represents an intercepted HTTP request and response transaction.
type TrafficRecord struct {
	ID                 string    `json:"id"`
	TargetID           string    `json:"target_id"`
	ScanID             string    `json:"scan_id,omitempty"`
	Method             string    `json:"method"`
	URL                string    `json:"url"`
	Path               string    `json:"path"`
	Host               string    `json:"host"`
	RequestHeaders     string    `json:"request_headers"` // JSON map[string][]string
	RequestBody        string    `json:"request_body"`
	ResponseStatusCode int       `json:"response_status_code"`
	ResponseHeaders    string    `json:"response_headers"` // JSON map[string][]string
	ResponseBody       string    `json:"response_body"`
	ContentType        string    `json:"content_type"`
	ContentLength      int64     `json:"content_length"`
	DurationMs         int64     `json:"duration_ms"`
	Category           string    `json:"category"`     // auth, api, upload, admin, search, general
	AuthContext        string    `json:"auth_context"` // unauthenticated, session_cookie, bearer_token, api_key
	Parameters         string    `json:"parameters"`   // JSON array of identified parameters
	CreatedAt          time.Time `json:"created_at"`
}
