package models

import "time"

// ScanStatus represents the lifecycle state of a tool run.
type ScanStatus string

const (
	StatusPending   ScanStatus = "PENDING"
	StatusRunning   ScanStatus = "RUNNING"
	StatusCompleted ScanStatus = "COMPLETED"
	StatusFailed    ScanStatus = "FAILED"
	StatusStopped   ScanStatus = "STOPPED"
)

// Scan represents a single execution session of a specific tool against a target.
type Scan struct {
	ID          string     `json:"id"`
	TargetID    string     `json:"target_id"`
	ToolName    string     `json:"tool_name"`
	Status      ScanStatus `json:"status"`
	Command     string     `json:"command"`
	Arguments   []string   `json:"arguments"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ErrorMsg    string     `json:"error_msg,omitempty"`
}
