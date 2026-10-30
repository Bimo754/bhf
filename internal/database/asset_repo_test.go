package database

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/diamond/bhf/internal/models"
)

func TestAssetRepoPriorityAndStats(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "test_assets.db")
	db, err := Open(tempDB)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := MigrateSchema(db); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	targetRepo := NewTargetRepo(db)
	scanRepo := NewScanRepo(db)
	assetRepo := NewAssetRepo(db)

	target, err := targetRepo.Create("Target", "target.com", "")
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	scan := &models.Scan{
		ID:        "scan-assets",
		TargetID:  target.ID,
		ToolName:  "httpx",
		Status:    models.StatusCompleted,
		Command:   "httpx",
		StartedAt: time.Now().UTC(),
	}
	if err := scanRepo.Create(scan); err != nil {
		t.Fatalf("failed to create scan: %v", err)
	}

	// Normal asset
	a1 := &models.Asset{
		ID:         "asset-1",
		TargetID:   target.ID,
		ScanID:     scan.ID,
		Host:       "static.target.com",
		URL:        "https://static.target.com",
		StatusCode: 200,
	}

	// High priority asset: non-standard port + admin keyword + 403 status code
	a2 := &models.Asset{
		ID:           "asset-2",
		TargetID:     target.ID,
		ScanID:       scan.ID,
		Host:         "admin-api.target.com",
		URL:          "https://admin-api.target.com:8443",
		Port:         8443,
		StatusCode:   403,
		Technologies: []string{"Spring Boot", "Swagger"},
	}

	if err := assetRepo.Upsert(a1); err != nil {
		t.Fatalf("failed to upsert a1: %v", err)
	}
	if err := assetRepo.Upsert(a2); err != nil {
		t.Fatalf("failed to upsert a2: %v", err)
	}

	stats, err := assetRepo.GetStats(target.ID)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.Total != 2 {
		t.Errorf("expected 2 assets, got %d", stats.Total)
	}
	if stats.HighPriority != 1 {
		t.Errorf("expected 1 high priority asset, got %d", stats.HighPriority)
	}

	// Verify priority ordering: a2 must be first
	assets, _, err := assetRepo.ListByTarget(target.ID, "", "", "priority", "desc", 10, 0)
	if err != nil {
		t.Fatalf("failed to list assets: %v", err)
	}
	if len(assets) < 2 || assets[0].Host != "admin-api.target.com" {
		t.Errorf("expected high-priority admin-api.target.com first, got %v", assets[0].Host)
	}

	// Verify reverse priority ordering: a1 must be first when order is asc
	assetsAsc, _, err := assetRepo.ListByTarget(target.ID, "", "", "priority", "asc", 10, 0)
	if err != nil {
		t.Fatalf("failed to list assets ascending: %v", err)
	}
	if len(assetsAsc) < 2 || assetsAsc[0].Host != "static.target.com" {
		t.Errorf("expected static.target.com first in ascending priority, got %v", assetsAsc[0].Host)
	}

	// Verify server-side filter: live filter returns 2xx/live assets
	filtered, totalFiltered, err := assetRepo.ListByTarget(target.ID, "", "errors", "priority", "desc", 10, 0)
	if err != nil {
		t.Fatalf("failed to list filtered assets: %v", err)
	}
	if totalFiltered != 1 || len(filtered) != 1 || filtered[0].StatusCode != 403 {
		t.Errorf("expected 1 error asset with 403, got %d (total %d)", len(filtered), totalFiltered)
	}
}
