package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// FindingRepo handles database operations for vulnerability and security findings.
type FindingRepo struct {
	db *sql.DB
}

// NewFindingRepo creates a new FindingRepo.
func NewFindingRepo(db *sql.DB) *FindingRepo {
	return &FindingRepo{db: db}
}

// Insert inserts a new finding into the database.
func (r *FindingRepo) Insert(finding *models.Finding) error {
	if finding.ID == "" {
		finding.ID = uuid.New().String()
	}
	if finding.CreatedAt.IsZero() {
		finding.CreatedAt = time.Now().UTC()
	}

	extractedJSON, err := json.Marshal(finding.ExtractedResults)
	if err != nil {
		extractedJSON = []byte("[]")
	}

	query := `INSERT INTO findings (
		id, target_id, scan_id, tool_name, severity, name, matched_at, description, extracted_results, curl_command, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.Exec(query,
		finding.ID,
		finding.TargetID,
		finding.ScanID,
		finding.ToolName,
		strings.ToLower(string(finding.Severity)),
		finding.Name,
		finding.MatchedAt,
		finding.Description,
		string(extractedJSON),
		finding.CurlCommand,
		finding.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert finding: %w", err)
	}

	return nil
}

// FindingStats holds aggregate counts of findings categorized by severity.
type FindingStats struct {
	Total    int `json:"total"`
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
}

// GetStats returns aggregate severity counts for a target.
func (r *FindingRepo) GetStats(targetID string) (FindingStats, error) {
	query := `SELECT 
		COUNT(*),
		COUNT(CASE WHEN LOWER(severity) = 'critical' THEN 1 END),
		COUNT(CASE WHEN LOWER(severity) = 'high' THEN 1 END),
		COUNT(CASE WHEN LOWER(severity) = 'medium' THEN 1 END),
		COUNT(CASE WHEN LOWER(severity) = 'low' THEN 1 END),
		COUNT(CASE WHEN LOWER(severity) = 'info' THEN 1 END)
	FROM findings WHERE target_id = ?`

	var stats FindingStats
	err := r.db.QueryRow(query, targetID).Scan(
		&stats.Total,
		&stats.Critical,
		&stats.High,
		&stats.Medium,
		&stats.Low,
		&stats.Info,
	)
	if err != nil {
		return stats, fmt.Errorf("failed to get finding stats: %w", err)
	}
	return stats, nil
}

// ListByTarget returns findings for a target with severity filtering, search, and pagination.
func (r *FindingRepo) ListByTarget(targetID string, severity string, search string, limit, offset int) ([]models.Finding, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var conditions []string
	var args []interface{}

	conditions = append(conditions, "target_id = ?")
	args = append(args, targetID)

	severity = strings.TrimSpace(strings.ToLower(severity))
	if severity != "" && severity != "all" {
		conditions = append(conditions, "LOWER(severity) = ?")
		args = append(args, severity)
	}

	if search != "" {
		searchPattern := "%" + search + "%"
		conditions = append(conditions, "(name LIKE ? OR matched_at LIKE ? OR description LIKE ?)")
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	whereClause := " WHERE " + strings.Join(conditions, " AND ")

	countQuery := "SELECT COUNT(*) FROM findings" + whereClause
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count findings: %w", err)
	}

	query := `SELECT id, target_id, scan_id, tool_name, severity, name, matched_at, description, extracted_results, curl_command, created_at
		FROM findings` + whereClause + `
		ORDER BY CASE LOWER(severity)
			WHEN 'critical' THEN 1
			WHEN 'high' THEN 2
			WHEN 'medium' THEN 3
			WHEN 'low' THEN 4
			WHEN 'info' THEN 5
			ELSE 6
		END, created_at DESC LIMIT ? OFFSET ?`

	queryArgs := append(args, limit, offset)
	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query findings: %w", err)
	}
	defer rows.Close()

	var findings []models.Finding
	for rows.Next() {
		var f models.Finding
		var extractedRaw string
		if err := rows.Scan(
			&f.ID, &f.TargetID, &f.ScanID, &f.ToolName, &f.Severity,
			&f.Name, &f.MatchedAt, &f.Description, &extractedRaw,
			&f.CurlCommand, &f.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan finding row: %w", err)
		}
		_ = json.Unmarshal([]byte(extractedRaw), &f.ExtractedResults)
		findings = append(findings, f)
	}

	return findings, total, nil
}

// Delete removes a finding by its ID.
func (r *FindingRepo) Delete(id string) error {
	query := `DELETE FROM findings WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete finding: %w", err)
	}
	return nil
}
