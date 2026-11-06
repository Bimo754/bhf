package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// HostRepo manages persistence operations for network hosts.
type HostRepo struct {
	db *sql.DB
}

// NewHostRepo creates a new HostRepo.
func NewHostRepo(db *sql.DB) *HostRepo {
	return &HostRepo{db: db}
}

// Upsert inserts or updates a discovered network host.
func (r *HostRepo) Upsert(h *models.Host) error {
	if h.ID == "" {
		h.ID = uuid.New().String()
	}
	if h.CreatedAt.IsZero() {
		h.CreatedAt = time.Now().UTC()
	}
	if h.Status == "" {
		h.Status = "up"
	}

	query := `INSERT INTO hosts (id, target_id, ip, hostname, os, mac, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(target_id, ip) DO UPDATE SET
			hostname = CASE WHEN excluded.hostname != '' THEN excluded.hostname ELSE hosts.hostname END,
			os = CASE WHEN excluded.os != '' THEN excluded.os ELSE hosts.os END,
			mac = CASE WHEN excluded.mac != '' THEN excluded.mac ELSE hosts.mac END,
			status = excluded.status`

	_, err := r.db.Exec(query,
		h.ID, h.TargetID, h.IP, h.Hostname, h.OS, h.MAC, h.Status, h.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert host: %w", err)
	}

	return nil
}

// GetByID retrieves a single host by ID.
func (r *HostRepo) GetByID(id string) (*models.Host, error) {
	query := `SELECT id, target_id, ip, hostname, os, mac, status, created_at FROM hosts WHERE id = ?`
	row := r.db.QueryRow(query, id)
	var h models.Host
	err := row.Scan(&h.ID, &h.TargetID, &h.IP, &h.Hostname, &h.OS, &h.MAC, &h.Status, &h.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %w", err)
	}
	return &h, nil
}

// GetByIP retrieves a single host by target ID and IP address.
func (r *HostRepo) GetByIP(targetID, ip string) (*models.Host, error) {
	query := `SELECT id, target_id, ip, hostname, os, mac, status, created_at FROM hosts WHERE target_id = ? AND ip = ?`
	row := r.db.QueryRow(query, targetID, ip)
	var h models.Host
	err := row.Scan(&h.ID, &h.TargetID, &h.IP, &h.Hostname, &h.OS, &h.MAC, &h.Status, &h.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get host by ip: %w", err)
	}
	return &h, nil
}

// ListByTarget returns paginated hosts for a target scope.
func (r *HostRepo) ListByTarget(targetID string, limit, offset int) ([]models.Host, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM hosts WHERE target_id = ?", targetID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count hosts: %w", err)
	}

	query := `SELECT id, target_id, ip, hostname, os, mac, status, created_at
		FROM hosts WHERE target_id = ?
		ORDER BY ip ASC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, targetID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list hosts: %w", err)
	}
	defer rows.Close()

	hosts := make([]models.Host, 0)
	for rows.Next() {
		var h models.Host
		if err := rows.Scan(&h.ID, &h.TargetID, &h.IP, &h.Hostname, &h.OS, &h.MAC, &h.Status, &h.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan host: %w", err)
		}
		hosts = append(hosts, h)
	}

	return hosts, total, nil
}
