package database

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/diamond/bhf/internal/models"
)

func TestEndpointRepoOperations(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "test_endpoints.db")
	db, err := Open(tempDB)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := MigrateSchema(db); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	targetRepo := NewTargetRepo(db)
	endpointRepo := NewEndpointRepo(db)

	target, err := targetRepo.Create("Target Org", "example.com", "")
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	now := time.Now().UTC()

	ep1 := &models.Endpoint{
		ID:         "ep-1",
		TargetID:   target.ID,
		Host:       "api.example.com",
		URL:        "https://api.example.com/v1/users",
		Path:       "/v1/users",
		Method:     "GET",
		StatusCode: 200,
		SourceTool: "katana",
		CreatedAt:  now,
	}

	ep2 := &models.Endpoint{
		ID:         "ep-2",
		TargetID:   target.ID,
		Host:       "api.example.com",
		URL:        "https://api.example.com/v1/auth",
		Path:       "/v1/auth",
		Method:     "POST",
		StatusCode: 401,
		SourceTool: "ffuf",
		CreatedAt:  now,
	}

	ep3 := &models.Endpoint{
		ID:         "ep-3",
		TargetID:   target.ID,
		Host:       "admin.example.com",
		URL:        "https://admin.example.com/dashboard",
		Path:       "/dashboard",
		Method:     "GET",
		StatusCode: 403,
		SourceTool: "gau",
		CreatedAt:  now,
	}

	if err := endpointRepo.Upsert(ep1); err != nil {
		t.Fatalf("failed to upsert ep1: %v", err)
	}
	if err := endpointRepo.Upsert(ep2); err != nil {
		t.Fatalf("failed to upsert ep2: %v", err)
	}
	if err := endpointRepo.Upsert(ep3); err != nil {
		t.Fatalf("failed to upsert ep3: %v", err)
	}

	// Conflict update test: upserting same URL updates status_code
	ep1Updated := &models.Endpoint{
		ID:         "ep-1-dup",
		TargetID:   target.ID,
		URL:        "https://api.example.com/v1/users",
		StatusCode: 204,
	}
	if err := endpointRepo.Upsert(ep1Updated); err != nil {
		t.Fatalf("failed to update ep1 on conflict: %v", err)
	}

	// Verify total count
	count, err := endpointRepo.Count(target.ID)
	if err != nil {
		t.Fatalf("failed to count endpoints: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 endpoints, got %d", count)
	}

	// Verify GetAllForTarget (sitemap hierarchy)
	all, err := endpointRepo.GetAllForTarget(target.ID)
	if err != nil {
		t.Fatalf("failed to get all endpoints: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 endpoints, got %d", len(all))
	}

	// Verify search
	filtered, total, err := endpointRepo.ListByTarget(target.ID, "auth", 10, 0)
	if err != nil {
		t.Fatalf("failed to list by search: %v", err)
	}
	if total != 1 || len(filtered) != 1 || filtered[0].Path != "/v1/auth" {
		t.Errorf("expected 1 endpoint matching auth, got %d", total)
	}
}
