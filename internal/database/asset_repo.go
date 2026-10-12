package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// AssetRepo handles database operations for discovered assets (subdomains, URLs, services).
type AssetRepo struct {
	db *sql.DB
}

// NewAssetRepo creates a new AssetRepo.
func NewAssetRepo(db *sql.DB) *AssetRepo {
	return &AssetRepo{db: db}
}

// Upsert adds or updates an asset, deduplicating by target_id, host, and port.
func (r *AssetRepo) Upsert(asset *models.Asset) error {
	if asset.ID == "" {
		asset.ID = uuid.New().String()
	}
	if asset.CreatedAt.IsZero() {
		asset.CreatedAt = time.Now().UTC()
	}

	techJSON, err := json.Marshal(asset.Technologies)
	if err != nil {
		techJSON = []byte("[]")
	}

	query := `INSERT INTO assets (id, target_id, scan_id, host, url, ip, port, status_code, title, technologies, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(target_id, host, port) DO UPDATE SET
			url = CASE WHEN excluded.url != '' THEN excluded.url ELSE assets.url END,
			ip = CASE WHEN excluded.ip != '' THEN excluded.ip ELSE assets.ip END,
			status_code = CASE WHEN excluded.status_code != 0 THEN excluded.status_code ELSE assets.status_code END,
			title = CASE WHEN excluded.title != '' THEN excluded.title ELSE assets.title END,
			technologies = CASE WHEN excluded.technologies != '[]' THEN excluded.technologies ELSE assets.technologies END`

	_, err = r.db.Exec(query,
		asset.ID, asset.TargetID, asset.ScanID, asset.Host, asset.URL,
		asset.IP, asset.Port, asset.StatusCode, asset.Title, string(techJSON), asset.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert asset: %w", err)
	}

	return nil
}

// ListByTarget returns assets for a target with pagination and optional search query.
func (r *AssetRepo) ListByTarget(targetID string, search string, limit, offset int) ([]models.Asset, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Count total matching
	countQuery := `SELECT COUNT(*) FROM assets WHERE target_id = ? AND (host LIKE ? OR title LIKE ? OR ip LIKE ?)`
	searchPattern := "%" + search + "%"
	var total int
	err := r.db.QueryRow(countQuery, targetID, searchPattern, searchPattern, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count assets: %w", err)
	}

	// Fetch paginated
	query := `SELECT id, target_id, scan_id, host, url, ip, port, status_code, title, technologies, created_at
		FROM assets 
		WHERE target_id = ? AND (host LIKE ? OR title LIKE ? OR ip LIKE ?)
		ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, targetID, searchPattern, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list assets: %w", err)
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var a models.Asset
		var techRaw string
		if err := rows.Scan(&a.ID, &a.TargetID, &a.ScanID, &a.Host, &a.URL, &a.IP, &a.Port, &a.StatusCode, &a.Title, &techRaw, &a.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan asset row: %w", err)
		}
		_ = json.Unmarshal([]byte(techRaw), &a.Technologies)
		assets = append(assets, a)
	}

	return assets, total, nil
}

// GetDistinctHosts returns all unique subdomains discovered for a target.
func (r *AssetRepo) GetDistinctHosts(targetID string) ([]string, error) {
	query := `SELECT DISTINCT host FROM assets WHERE target_id = ? AND host != '' ORDER BY host ASC`
	rows, err := r.db.Query(query, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get distinct hosts: %w", err)
	}
	defer rows.Close()

	var hosts []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err == nil && h != "" {
			hosts = append(hosts, h)
		}
	}
	return hosts, nil
}

// GetLiveURLs returns all active web URLs discovered for a target.
func (r *AssetRepo) GetLiveURLs(targetID string) ([]string, error) {
	query := `SELECT DISTINCT url FROM assets WHERE target_id = ? AND url != '' ORDER BY url ASC`
	rows, err := r.db.Query(query, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get live URLs: %w", err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err == nil && u != "" {
			urls = append(urls, u)
		}
	}
	return urls, nil
}
