package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/diamond/bhf/internal/models"
)

// ScanRepo handles database operations for tool execution sessions.
type ScanRepo struct {
	db *sql.DB
}

// NewScanRepo creates a new ScanRepo.
func NewScanRepo(db *sql.DB) *ScanRepo {
	return &ScanRepo{db: db}
}

// Create records the launch of a new tool scan.
func (r *ScanRepo) Create(scan *models.Scan) error {
	argsJSON, err := json.Marshal(scan.Arguments)
	if err != nil {
		argsJSON = []byte("[]")
	}

	query := `INSERT INTO scans (id, target_id, tool_name, status, command, arguments, started_at, error_msg)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = r.db.Exec(query, scan.ID, scan.TargetID, scan.ToolName, scan.Status, scan.Command, string(argsJSON), scan.StartedAt, scan.ErrorMsg)
	if err != nil {
		return fmt.Errorf("failed to insert scan: %w", err)
	}
	return nil
}

// UpdateStatus updates the execution state, completion time, and any error message of a scan.
func (r *ScanRepo) UpdateStatus(id string, status models.ScanStatus, completedAt *time.Time, errMsg string) error {
	query := `UPDATE scans SET status = ?, completed_at = ?, error_msg = ? WHERE id = ?`
	_, err := r.db.Exec(query, status, completedAt, errMsg, id)
	if err != nil {
		return fmt.Errorf("failed to update scan status: %w", err)
	}
	return nil
}

// GetByID returns a single scan record.
func (r *ScanRepo) GetByID(id string) (*models.Scan, error) {
	query := `SELECT id, target_id, tool_name, status, command, arguments, started_at, completed_at, error_msg FROM scans WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var s models.Scan
	var argsRaw string
	if err := row.Scan(&s.ID, &s.TargetID, &s.ToolName, &s.Status, &s.Command, &argsRaw, &s.StartedAt, &s.CompletedAt, &s.ErrorMsg); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get scan: %w", err)
	}

	_ = json.Unmarshal([]byte(argsRaw), &s.Arguments)
	return &s, nil
}

// ListByTarget returns scans for a specific target ordered by started_at descending.
func (r *ScanRepo) ListByTarget(targetID string) ([]models.Scan, error) {
	query := `SELECT id, target_id, tool_name, status, command, arguments, started_at, completed_at, error_msg 
		FROM scans WHERE target_id = ? ORDER BY started_at DESC`
	rows, err := r.db.Query(query, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to list scans: %w", err)
	}
	defer rows.Close()

	var scans []models.Scan
	for rows.Next() {
		var s models.Scan
		var argsRaw string
		if err := rows.Scan(&s.ID, &s.TargetID, &s.ToolName, &s.Status, &s.Command, &argsRaw, &s.StartedAt, &s.CompletedAt, &s.ErrorMsg); err != nil {
			return nil, fmt.Errorf("failed to scan scan row: %w", err)
		}
		_ = json.Unmarshal([]byte(argsRaw), &s.Arguments)
		scans = append(scans, s)
	}

	return scans, nil
}
