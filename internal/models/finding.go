package models

import "time"

// Severity represents finding severity levels.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Finding represents an identified vulnerability, exposure, or security issue.
type Finding struct {
	ID               string    `json:"id"`
	TargetID         string    `json:"target_id"`
	ScanID           string    `json:"scan_id"`
	ToolName         string    `json:"tool_name"`
	Severity         Severity  `json:"severity"`
	Name             string    `json:"name"`
	MatchedAt        string    `json:"matched_at"`
	Description      string    `json:"description,omitempty"`
	ExtractedResults []string  `json:"extracted_results,omitempty"`
	CurlCommand      string    `json:"curl_command,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}
