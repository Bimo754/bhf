package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// TargetRepo handles database operations for target scopes.
type TargetRepo struct {
	db *sql.DB
}

// NewTargetRepo creates a new TargetRepo.
func NewTargetRepo(db *sql.DB) *TargetRepo {
	return &TargetRepo{db: db}
}

// Create inserts a new target into the database.
func (r *TargetRepo) Create(name, domain, notes string) (*models.Target, error) {
	now := time.Now().UTC()
	target := &models.Target{
		ID:        uuid.New().String(),
		Name:      name,
		Domain:    domain,
		Notes:     notes,
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `INSERT INTO targets (id, name, domain, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, target.ID, target.Name, target.Domain, target.Notes, target.CreatedAt, target.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert target: %w", err)
	}

	return target, nil
}

// List returns all targets ordered by creation time descending.
func (r *TargetRepo) List() ([]models.Target, error) {
	query := `SELECT id, name, domain, notes, created_at, updated_at FROM targets ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list targets: %w", err)
	}
	defer rows.Close()

	var targets []models.Target
	for rows.Next() {
		var t models.Target
		if err := rows.Scan(&t.ID, &t.Name, &t.Domain, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan target: %w", err)
		}
		targets = append(targets, t)
	}

	return targets, nil
}

// GetByID returns a single target by ID.
func (r *TargetRepo) GetByID(id string) (*models.Target, error) {
	query := `SELECT id, name, domain, notes, created_at, updated_at FROM targets WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var t models.Target
	if err := row.Scan(&t.ID, &t.Name, &t.Domain, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get target: %w", err)
	}

	return &t, nil
}

// Delete removes a target and cascades to associated scans and assets.
func (r *TargetRepo) Delete(id string) error {
	query := `DELETE FROM targets WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete target: %w", err)
	}
	return nil
}
