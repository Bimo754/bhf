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

// AssetRepo handles database operations for discovered assets (subdomains, URLs, services).
type AssetRepo struct {
	db *sql.DB
}

// NewAssetRepo creates a new AssetRepo.
func NewAssetRepo(db *sql.DB) *AssetRepo {
	return &AssetRepo{db: db}
}

// CalculatePriorityScore computes a heuristic priority score to highlight high-value bug bounty targets.
func CalculatePriorityScore(host, urlStr, title string, port, statusCode int, technologies []string) int {
	score := 0

	// 1. Non-standard ports (e.g. 8080, 8443, 9000, 3000, 5000, 8888)
	if port > 0 && port != 80 && port != 443 {
		score += 30
	}

	// 2. 401/403 status codes (prime candidates for access control and header bypass testing)
	if statusCode == 401 || statusCode == 403 {
		score += 20
	} else if statusCode >= 500 {
		score += 15
	}

	// 3. Sensitive subdomain / path naming patterns
	sensitiveKeywords := []string{
		"admin", "internal", "stage", "staging", "dev", "test",
		"corp", "vpn", "auth", "portal", "api", "dashboard",
		"jenkins", "jira", "grafana", "sso", "login",
	}
	combined := strings.ToLower(host + " " + urlStr + " " + title)
	for _, kw := range sensitiveKeywords {
		if strings.Contains(combined, kw) {
			score += 15
			break
		}
	}

	// 4. High-value application frameworks and exposed technologies
	highValueTech := []string{
		"spring", "laravel", "jenkins", "jira", "grafana",
		"swagger", "graphql", "webpack", "docker", "kubernetes",
		"php", "wordpress", "django", "rails", "node",
	}
	for _, t := range technologies {
		tLower := strings.ToLower(t)
		for _, hvt := range highValueTech {
			if strings.Contains(tLower, hvt) {
				score += 20
				break
			}
		}
	}

	return score
}

// Upsert adds or updates an asset, deduplicating by target_id, host, and port.
func (r *AssetRepo) Upsert(asset *models.Asset) error {
	if asset.ID == "" {
		asset.ID = uuid.New().String()
	}
	if asset.CreatedAt.IsZero() {
		asset.CreatedAt = time.Now().UTC()
	}

	if asset.PriorityScore == 0 {
		asset.PriorityScore = CalculatePriorityScore(
			asset.Host, asset.URL, asset.Title, asset.Port, asset.StatusCode, asset.Technologies,
		)
	}

	techJSON, err := json.Marshal(asset.Technologies)
	if err != nil {
		techJSON = []byte("[]")
	}

	query := `INSERT INTO assets (id, target_id, scan_id, host, url, ip, port, status_code, title, technologies, priority_score, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(target_id, host, port) DO UPDATE SET
			url = CASE WHEN excluded.url != '' THEN excluded.url ELSE assets.url END,
			ip = CASE WHEN excluded.ip != '' THEN excluded.ip ELSE assets.ip END,
			status_code = CASE WHEN excluded.status_code != 0 THEN excluded.status_code ELSE assets.status_code END,
			title = CASE WHEN excluded.title != '' THEN excluded.title ELSE assets.title END,
			technologies = CASE WHEN excluded.technologies != '[]' THEN excluded.technologies ELSE assets.technologies END,
			priority_score = CASE WHEN excluded.priority_score > assets.priority_score THEN excluded.priority_score ELSE assets.priority_score END`

	_, err = r.db.Exec(query,
		asset.ID, asset.TargetID, asset.ScanID, asset.Host, asset.URL,
		asset.IP, asset.Port, asset.StatusCode, asset.Title, string(techJSON), asset.PriorityScore, asset.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert asset: %w", err)
	}

	return nil
}

// AssetStats holds aggregate counts for different asset classifications.
type AssetStats struct {
	Total        int `json:"total"`
	Live         int `json:"live"`
	Code2xx      int `json:"code_2xx"`
	Code3xx      int `json:"code_3xx"`
	Code4xx5xx   int `json:"code_4xx_5xx"`
	HighPriority int `json:"high_priority"`
	NewAssets    int `json:"new_assets"`
}

// GetStats returns aggregate classification counts for a target scope.
func (r *AssetRepo) GetStats(targetID string) (AssetStats, error) {
	query := `SELECT 
		COUNT(*),
		COUNT(CASE WHEN status_code > 0 OR url != '' THEN 1 END),
		COUNT(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 END),
		COUNT(CASE WHEN status_code >= 300 AND status_code < 400 THEN 1 END),
		COUNT(CASE WHEN status_code >= 400 THEN 1 END),
		COUNT(CASE WHEN priority_score >= 30 THEN 1 END),
		COUNT(CASE WHEN created_at >= datetime('now', '-24 hours') THEN 1 END)
	FROM assets WHERE target_id = ?`

	var stats AssetStats
	err := r.db.QueryRow(query, targetID).Scan(
		&stats.Total,
		&stats.Live,
		&stats.Code2xx,
		&stats.Code3xx,
		&stats.Code4xx5xx,
		&stats.HighPriority,
		&stats.NewAssets,
	)
	if err != nil {
		return stats, fmt.Errorf("failed to get asset stats: %w", err)
	}
	return stats, nil
}

// ListByTarget returns assets for a target with pagination, search, filter, and sorting.
func (r *AssetRepo) ListByTarget(targetID string, search string, filter string, sortBy string, order string, limit, offset int) ([]models.Asset, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	filterClause := ""
	switch filter {
	case "live":
		filterClause = " AND (status_code > 0 OR url != '')"
	case "2xx":
		filterClause = " AND (status_code >= 200 AND status_code < 300)"
	case "3xx":
		filterClause = " AND (status_code >= 300 AND status_code < 400)"
	case "errors":
		filterClause = " AND (status_code >= 400)"
	case "priority":
		filterClause = " AND (priority_score >= 30)"
	case "new":
		filterClause = " AND (created_at >= datetime('now', '-24 hours'))"
	}

	searchPattern := "%" + search + "%"

	// Count total matching
	countQuery := `SELECT COUNT(*) FROM assets 
		WHERE target_id = ? AND (host LIKE ? OR title LIKE ? OR ip LIKE ? OR url LIKE ?)` + filterClause
	var total int
	err := r.db.QueryRow(countQuery, targetID, searchPattern, searchPattern, searchPattern, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count assets: %w", err)
	}

	orderDir := "DESC"
	if strings.ToLower(order) == "asc" {
		orderDir = "ASC"
	}

	orderClause := "priority_score DESC, created_at DESC"
	switch strings.ToLower(sortBy) {
	case "priority":
		orderClause = fmt.Sprintf("priority_score %s, created_at DESC", orderDir)
	case "host":
		orderClause = fmt.Sprintf("host %s", orderDir)
	case "status":
		orderClause = fmt.Sprintf("status_code %s, priority_score DESC", orderDir)
	case "port":
		orderClause = fmt.Sprintf("port %s", orderDir)
	case "created_at":
		orderClause = fmt.Sprintf("created_at %s", orderDir)
	default:
		if strings.ToLower(order) == "asc" {
			orderClause = "priority_score ASC, created_at DESC"
		}
	}

	// Fetch paginated ordered by dynamic order clause
	query := `SELECT id, target_id, scan_id, host, url, ip, port, status_code, title, technologies, priority_score, created_at
		FROM assets 
		WHERE target_id = ? AND (host LIKE ? OR title LIKE ? OR ip LIKE ? OR url LIKE ?)` + filterClause + `
		ORDER BY ` + orderClause + ` LIMIT ? OFFSET ?`

	rows, err := r.db.Query(query, targetID, searchPattern, searchPattern, searchPattern, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list assets: %w", err)
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var a models.Asset
		var techRaw string
		if err := rows.Scan(&a.ID, &a.TargetID, &a.ScanID, &a.Host, &a.URL, &a.IP, &a.Port, &a.StatusCode, &a.Title, &techRaw, &a.PriorityScore, &a.CreatedAt); err != nil {
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
