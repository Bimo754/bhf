package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// ServiceRepo manages persistence operations for discovered network services.
type ServiceRepo struct {
	db *sql.DB
}

// NewServiceRepo creates a new ServiceRepo.
func NewServiceRepo(db *sql.DB) *ServiceRepo {
	return &ServiceRepo{db: db}
}

// Upsert inserts or updates a discovered network service on a host.
func (r *ServiceRepo) Upsert(s *models.NetworkService) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now().UTC()
	}
	if s.Protocol == "" {
		s.Protocol = "tcp"
	}
	if s.Status == "" {
		s.Status = "open"
	}

	query := `INSERT INTO services (id, host_id, target_id, port, protocol, name, product, version, banner, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(host_id, port, protocol) DO UPDATE SET
			name = CASE WHEN excluded.name != '' THEN excluded.name ELSE services.name END,
			product = CASE WHEN excluded.product != '' THEN excluded.product ELSE services.product END,
			version = CASE WHEN excluded.version != '' THEN excluded.version ELSE services.version END,
			banner = CASE WHEN excluded.banner != '' THEN excluded.banner ELSE services.banner END,
			status = excluded.status`

	_, err := r.db.Exec(query,
		s.ID, s.HostID, s.TargetID, s.Port, s.Protocol,
		s.Name, s.Product, s.Version, s.Banner, s.Status, s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert network service: %w", err)
	}

	return nil
}

// ListByHost returns all open network services running on a specific host.
func (r *ServiceRepo) ListByHost(hostID string) ([]models.NetworkService, error) {
	query := `SELECT id, host_id, target_id, port, protocol, name, product, version, banner, status, created_at
		FROM services WHERE host_id = ? ORDER BY port ASC`

	rows, err := r.db.Query(query, hostID)
	if err != nil {
		return nil, fmt.Errorf("failed to list services for host: %w", err)
	}
	defer rows.Close()

	services := make([]models.NetworkService, 0)
	for rows.Next() {
		var s models.NetworkService
		if err := rows.Scan(&s.ID, &s.HostID, &s.TargetID, &s.Port, &s.Protocol, &s.Name, &s.Product, &s.Version, &s.Banner, &s.Status, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, s)
	}

	return services, nil
}

// ListByTarget returns paginated network services across all hosts in a target scope.
func (r *ServiceRepo) ListByTarget(targetID string, limit, offset int) ([]models.NetworkService, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM services WHERE target_id = ?", targetID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count services: %w", err)
	}

	query := `SELECT id, host_id, target_id, port, protocol, name, product, version, banner, status, created_at
		FROM services WHERE target_id = ?
		ORDER BY port ASC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, targetID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list services: %w", err)
	}
	defer rows.Close()

	services := make([]models.NetworkService, 0)
	for rows.Next() {
		var s models.NetworkService
		if err := rows.Scan(&s.ID, &s.HostID, &s.TargetID, &s.Port, &s.Protocol, &s.Name, &s.Product, &s.Version, &s.Banner, &s.Status, &s.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, s)
	}

	return services, total, nil
}

// GetHTTPWebServices returns all discovered services running HTTP or HTTPS for automated web assessment handoff.
func (r *ServiceRepo) GetHTTPWebServices(targetID string) ([]models.NetworkService, error) {
	query := `SELECT id, host_id, target_id, port, protocol, name, product, version, banner, status, created_at
		FROM services WHERE target_id = ? AND (
			LOWER(name) IN ('http', 'https') OR
			port IN (80, 443, 8080, 8443, 8000, 9000, 3000, 5000)
		) ORDER BY port ASC`

	rows, err := r.db.Query(query, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to query web services: %w", err)
	}
	defer rows.Close()

	webServices := make([]models.NetworkService, 0)
	for rows.Next() {
		var s models.NetworkService
		if err := rows.Scan(&s.ID, &s.HostID, &s.TargetID, &s.Port, &s.Protocol, &s.Name, &s.Product, &s.Version, &s.Banner, &s.Status, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan web service: %w", err)
		}
		webServices = append(webServices, s)
	}

	return webServices, nil
}
