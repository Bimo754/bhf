package models

import "time"

// LogStream distinguishes between stdout and stderr.
type LogStream string

const (
	StreamStdout LogStream = "stdout"
	StreamStderr LogStream = "stderr"
	StreamSystem LogStream = "system"
)

// LogEntry represents an individual streamed line of output from a running tool.
type LogEntry struct {
	ID        int64     `json:"id,omitempty"`
	ScanID    string    `json:"scan_id"`
	Stream    LogStream `json:"stream"`
	Line      string    `json:"line"`
	Timestamp time.Time `json:"timestamp"`
}
