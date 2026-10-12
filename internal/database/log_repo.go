package database

import (
	"database/sql"
	"fmt"

	"github.com/diamond/bhf/internal/models"
)

// LogRepo handles persisting and querying execution logs.
type LogRepo struct {
	db *sql.DB
}

// NewLogRepo creates a new LogRepo.
func NewLogRepo(db *sql.DB) *LogRepo {
	return &LogRepo{db: db}
}

// Append inserts a single log entry.
func (r *LogRepo) Append(entry *models.LogEntry) error {
	query := `INSERT INTO logs (scan_id, stream, line, timestamp) VALUES (?, ?, ?, ?)`
	res, err := r.db.Exec(query, entry.ScanID, entry.Stream, entry.Line, entry.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to append log: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil {
		entry.ID = id
	}
	return nil
}

// GetByScan returns all persisted log entries for a given scan session.
func (r *LogRepo) GetByScan(scanID string) ([]models.LogEntry, error) {
	query := `SELECT id, scan_id, stream, line, timestamp FROM logs WHERE scan_id = ? ORDER BY id ASC`
	rows, err := r.db.Query(query, scanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for scan: %w", err)
	}
	defer rows.Close()

	var logs []models.LogEntry
	for rows.Next() {
		var l models.LogEntry
		if err := rows.Scan(&l.ID, &l.ScanID, &l.Stream, &l.Line, &l.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan log row: %w", err)
		}
		logs = append(logs, l)
	}

	return logs, nil
}
