package database

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/diamond/bhf/internal/models"
)

func TestFindingRepo(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "test_findings.db")
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
	findingRepo := NewFindingRepo(db)

	target, err := targetRepo.Create("Example", "example.com", "")
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	scan := &models.Scan{
		ID:        "scan-1",
		TargetID:  target.ID,
		ToolName:  "nuclei",
		Status:    models.StatusCompleted,
		Command:   "nuclei -u example.com",
		StartedAt: time.Now().UTC(),
	}
	if err := scanRepo.Create(scan); err != nil {
		t.Fatalf("failed to create scan: %v", err)
	}

	f1 := &models.Finding{
		ID:          "finding-1",
		TargetID:    target.ID,
		ScanID:      scan.ID,
		ToolName:    "nuclei",
		Severity:    models.SeverityHigh,
		Name:        "Git Config Disclosure",
		MatchedAt:   "https://example.com/.git/config",
		CurlCommand: "curl -i https://example.com/.git/config",
	}

	f2 := &models.Finding{
		ID:          "finding-2",
		TargetID:    target.ID,
		ScanID:      scan.ID,
		ToolName:    "nuclei",
		Severity:    models.SeverityCritical,
		Name:        "Remote Code Execution",
		MatchedAt:   "https://example.com/api/eval",
		CurlCommand: "curl -X POST https://example.com/api/eval",
	}

	if err := findingRepo.Insert(f1); err != nil {
		t.Fatalf("failed to insert finding 1: %v", err)
	}
	if err := findingRepo.Insert(f2); err != nil {
		t.Fatalf("failed to insert finding 2: %v", err)
	}

	// Test stats
	stats, err := findingRepo.GetStats(target.ID)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.Total != 2 || stats.Critical != 1 || stats.High != 1 {
		t.Errorf("unexpected stats: %+v", stats)
	}

	// Test list and ordering (Critical must precede High)
	list, total, err := findingRepo.ListByTarget(target.ID, "", "", 10, 0)
	if err != nil {
		t.Fatalf("failed to list findings: %v", err)
	}
	if total != 2 {
		t.Errorf("expected 2 findings, got %d", total)
	}
	if len(list) < 2 || list[0].Severity != models.SeverityCritical {
		t.Errorf("expected first finding to be Critical, got %v", list[0].Severity)
	}
}
