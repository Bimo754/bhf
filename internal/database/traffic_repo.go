package database

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// TrafficRepo handles persistence and query operations for intercepted HTTP transactions.
type TrafficRepo struct {
	db *sql.DB
}

// NewTrafficRepo creates a new TrafficRepo instance.
func NewTrafficRepo(db *sql.DB) *TrafficRepo {
	return &TrafficRepo{db: db}
}

// Insert saves an intercepted HTTP traffic transaction.
func (r *TrafficRepo) Insert(record *models.TrafficRecord) error {
	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now().UTC()
	}
	if record.Category == "" {
		record.Category = "general"
	}
	if record.RequestHeaders == "" {
		record.RequestHeaders = "{}"
	}
	if record.ResponseHeaders == "" {
		record.ResponseHeaders = "{}"
	}
	if record.Parameters == "" {
		record.Parameters = "[]"
	}
	if record.Host == "" || record.Path == "" {
		if u, err := url.Parse(record.URL); err == nil {
			if record.Host == "" {
				record.Host = u.Hostname()
			}
			if record.Path == "" {
				record.Path = u.Path
				if record.Path == "" {
					record.Path = "/"
				}
				if u.RawQuery != "" {
					record.Path += "?" + u.RawQuery
				}
			}
		}
	}

	query := `INSERT INTO traffic_records (
		id, target_id, scan_id, method, url, path, host,
		request_headers, request_body, response_status_code, response_headers,
		response_body, content_type, content_length, duration_ms, category,
		auth_context, parameters, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query,
		record.ID, record.TargetID, record.ScanID, record.Method, record.URL,
		record.Path, record.Host, record.RequestHeaders, record.RequestBody,
		record.ResponseStatusCode, record.ResponseHeaders, record.ResponseBody,
		record.ContentType, record.ContentLength, record.DurationMs, record.Category,
		record.AuthContext, record.Parameters, record.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert traffic record: %w", err)
	}

	return nil
}

// GetByID retrieves a single traffic record by its ID.
func (r *TrafficRepo) GetByID(id string) (*models.TrafficRecord, error) {
	query := `SELECT id, target_id, scan_id, method, url, path, host,
		request_headers, request_body, response_status_code, response_headers,
		response_body, content_type, content_length, duration_ms, category,
		auth_context, parameters, created_at
		FROM traffic_records WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var rec models.TrafficRecord
	err := row.Scan(
		&rec.ID, &rec.TargetID, &rec.ScanID, &rec.Method, &rec.URL,
		&rec.Path, &rec.Host, &rec.RequestHeaders, &rec.RequestBody,
		&rec.ResponseStatusCode, &rec.ResponseHeaders, &rec.ResponseBody,
		&rec.ContentType, &rec.ContentLength, &rec.DurationMs, &rec.Category,
		&rec.AuthContext, &rec.Parameters, &rec.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get traffic record: %w", err)
	}

	return &rec, nil
}

// ListByTarget returns paginated traffic records for a target with optional category, method, and search filtering.
func (r *TrafficRepo) ListByTarget(targetID string, category string, method string, search string, limit, offset int) ([]models.TrafficRecord, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	whereClauses := []string{"target_id = ?"}
	args := []any{targetID}

	if category != "" && category != "all" {
		whereClauses = append(whereClauses, "category = ?")
		args = append(args, category)
	}
	if method != "" && method != "all" {
		whereClauses = append(whereClauses, "method = ?")
		args = append(args, strings.ToUpper(method))
	}
	if search != "" {
		whereClauses = append(whereClauses, "(url LIKE ? OR path LIKE ? OR host LIKE ?)")
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM traffic_records WHERE %s", whereSQL)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count traffic records: %w", err)
	}

	selectQuery := fmt.Sprintf(`SELECT id, target_id, scan_id, method, url, path, host,
		request_headers, request_body, response_status_code, response_headers,
		response_body, content_type, content_length, duration_ms, category,
		auth_context, parameters, created_at
		FROM traffic_records WHERE %s
		ORDER BY created_at DESC LIMIT ? OFFSET ?`, whereSQL)

	selectArgs := append(args, limit, offset)
	rows, err := r.db.Query(selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list traffic records: %w", err)
	}
	defer rows.Close()

	records := make([]models.TrafficRecord, 0)
	for rows.Next() {
		var rec models.TrafficRecord
		err := rows.Scan(
			&rec.ID, &rec.TargetID, &rec.ScanID, &rec.Method, &rec.URL,
			&rec.Path, &rec.Host, &rec.RequestHeaders, &rec.RequestBody,
			&rec.ResponseStatusCode, &rec.ResponseHeaders, &rec.ResponseBody,
			&rec.ContentType, &rec.ContentLength, &rec.DurationMs, &rec.Category,
			&rec.AuthContext, &rec.Parameters, &rec.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan traffic record: %w", err)
		}
		records = append(records, rec)
	}

	return records, total, nil
}

// CountByTarget returns the total number of traffic records for a target.
func (r *TrafficRepo) CountByTarget(targetID string) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM traffic_records WHERE target_id = ?", targetID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count traffic records: %w", err)
	}
	return count, nil
}

// GetCategoryStats returns a map of category names to their respective counts for a target.
func (r *TrafficRepo) GetCategoryStats(targetID string) (map[string]int, error) {
	query := `SELECT category, COUNT(*) FROM traffic_records WHERE target_id = ? GROUP BY category`
	rows, err := r.db.Query(query, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to query traffic category stats: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("failed to scan category stat: %w", err)
		}
		stats[category] = count
	}

	return stats, nil
}

// DeleteByTarget removes all traffic records associated with a target.
func (r *TrafficRepo) DeleteByTarget(targetID string) error {
	_, err := r.db.Exec("DELETE FROM traffic_records WHERE target_id = ?", targetID)
	if err != nil {
		return fmt.Errorf("failed to delete traffic records: %w", err)
	}
	return nil
}
