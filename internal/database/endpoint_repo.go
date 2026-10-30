package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// EndpointRepo handles database operations for crawled routes and web endpoints.
type EndpointRepo struct {
	db *sql.DB
}

// NewEndpointRepo creates a new EndpointRepo.
func NewEndpointRepo(db *sql.DB) *EndpointRepo {
	return &EndpointRepo{db: db}
}

// Upsert records a discovered endpoint, updating attributes on conflict.
func (r *EndpointRepo) Upsert(ep *models.Endpoint) error {
	if ep.ID == "" {
		ep.ID = uuid.New().String()
	}
	if ep.CreatedAt.IsZero() {
		ep.CreatedAt = time.Now().UTC()
	}
	if ep.Method == "" {
		ep.Method = "GET"
	}
	if ep.Host == "" || ep.Path == "" {
		if u, err := url.Parse(ep.URL); err == nil {
			if ep.Host == "" {
				ep.Host = u.Hostname()
			}
			if ep.Path == "" {
				ep.Path = u.Path
				if ep.Path == "" {
					ep.Path = "/"
				}
				if u.RawQuery != "" {
					ep.Path += "?" + u.RawQuery
				}
			}
		}
	}

	query := `INSERT INTO endpoints (id, target_id, scan_id, host, url, path, method, status_code, content_length, title, source_tool, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(target_id, url) DO UPDATE SET
			status_code = CASE WHEN excluded.status_code != 0 THEN excluded.status_code ELSE endpoints.status_code END,
			content_length = CASE WHEN excluded.content_length != 0 THEN excluded.content_length ELSE endpoints.content_length END,
			title = CASE WHEN excluded.title != '' THEN excluded.title ELSE endpoints.title END,
			source_tool = CASE WHEN excluded.source_tool != '' THEN excluded.source_tool ELSE endpoints.source_tool END`

	_, err := r.db.Exec(query,
		ep.ID, ep.TargetID, ep.ScanID, ep.Host, ep.URL, ep.Path,
		ep.Method, ep.StatusCode, ep.ContentLength, ep.Title, ep.SourceTool, ep.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert endpoint: %w", err)
	}

	return nil
}

// ListByTarget returns paginated endpoints for a target scope.
func (r *EndpointRepo) ListByTarget(targetID string, search string, limit, offset int) ([]models.Endpoint, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	searchPattern := "%" + search + "%"

	countQuery := `SELECT COUNT(*) FROM endpoints WHERE target_id = ? AND (url LIKE ? OR host LIKE ? OR path LIKE ? OR title LIKE ?)`
	var total int
	err := r.db.QueryRow(countQuery, targetID, searchPattern, searchPattern, searchPattern, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count endpoints: %w", err)
	}

	query := `SELECT id, target_id, scan_id, host, url, path, method, status_code, content_length, title, source_tool, created_at
		FROM endpoints
		WHERE target_id = ? AND (url LIKE ? OR host LIKE ? OR path LIKE ? OR title LIKE ?)
		ORDER BY created_at DESC, url ASC
		LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, targetID, searchPattern, searchPattern, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list endpoints: %w", err)
	}
	defer rows.Close()

	var endpoints []models.Endpoint
	for rows.Next() {
		var ep models.Endpoint
		if err := rows.Scan(
			&ep.ID, &ep.TargetID, &ep.ScanID, &ep.Host, &ep.URL, &ep.Path,
			&ep.Method, &ep.StatusCode, &ep.ContentLength, &ep.Title, &ep.SourceTool, &ep.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan endpoint row: %w", err)
		}
		endpoints = append(endpoints, ep)
	}

	return endpoints, total, nil
}

// GetAllForTarget returns all endpoints for a target without pagination, specifically for building sitemaps.
func (r *EndpointRepo) GetAllForTarget(targetID string) ([]models.Endpoint, error) {
	query := `SELECT id, target_id, scan_id, host, url, path, method, status_code, content_length, title, source_tool, created_at
		FROM endpoints
		WHERE target_id = ?
		ORDER BY host ASC, path ASC`

	rows, err := r.db.Query(query, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all endpoints: %w", err)
	}
	defer rows.Close()

	var endpoints []models.Endpoint
	for rows.Next() {
		var ep models.Endpoint
		if err := rows.Scan(
			&ep.ID, &ep.TargetID, &ep.ScanID, &ep.Host, &ep.URL, &ep.Path,
			&ep.Method, &ep.StatusCode, &ep.ContentLength, &ep.Title, &ep.SourceTool, &ep.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan endpoint: %w", err)
		}
		endpoints = append(endpoints, ep)
	}

	return endpoints, nil
}

// Count returns the total number of endpoints registered for a target.
func (r *EndpointRepo) Count(targetID string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM endpoints WHERE target_id = ?`, targetID).Scan(&count)
	return count, err
}
