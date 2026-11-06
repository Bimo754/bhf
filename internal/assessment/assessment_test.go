package assessment_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/diamond/bhf/internal/assessment"
	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
)

func TestAuthSecurityRule(t *testing.T) {
	rule := assessment.NewAuthSecurityRule()
	target := &models.Target{ID: "tgt-1", Domain: "example.com"}

	traffic := []models.TrafficRecord{
		{
			URL:                "https://example.com/login?username=alice&password=secret123",
			Method:             "GET",
			ResponseStatusCode: 200,
			ResponseHeaders:    `{"Set-Cookie":["session_id=abc; Path=/"]}`,
		},
	}

	ctx := context.Background()
	findings, err := rule.Evaluate(ctx, target, traffic, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) < 2 {
		t.Fatalf("expected at least 2 findings (query credential + missing cookie flags), got %d", len(findings))
	}

	foundCredExposure := false
	foundCookieFlags := false
	for _, f := range findings {
		if strings.Contains(f.Name, "Sensitive Credentials Exposed in URL") {
			foundCredExposure = true
		}
		if strings.Contains(f.Name, "Session Cookie Missing Security Flags") {
			foundCookieFlags = true
		}
	}

	if !foundCredExposure {
		t.Errorf("missing credential exposure finding")
	}
	if !foundCookieFlags {
		t.Errorf("missing cookie security flags finding")
	}
}

func TestIDORSecurityRule(t *testing.T) {
	// Mock server that returns other users' data when ID is incremented
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("user_id")
		if id == "101" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":101,"name":"Alice"}`))
			return
		}
		if id == "102" {
			// IDOR vulnerability: returns Bob's record to Alice's request
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id":102,"name":"Bob","ssn":"111-22-3333"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	rule := assessment.NewIDORSecurityRule()
	target := &models.Target{ID: "tgt-1", Domain: "example.com"}

	traffic := []models.TrafficRecord{
		{
			URL:                mockServer.URL + "/api/user?user_id=101",
			Method:             "GET",
			ResponseStatusCode: 200,
		},
	}

	ctx := context.Background()
	findings, err := rule.Evaluate(ctx, target, traffic, mockServer.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 IDOR finding, got %d", len(findings))
	}

	finding := findings[0]
	if !strings.Contains(finding.Name, "Broken Object-Level Authorization") {
		t.Errorf("unexpected finding name: %s", finding.Name)
	}
	if finding.Confidence != assessment.ConfidenceConfirmed {
		t.Errorf("expected CONFIRMED confidence, got %s", finding.Confidence)
	}
}

func TestInputValidationRule(t *testing.T) {
	// Mock server that reflects user input without encoding
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		// Vulnerable reflection
		w.Write([]byte("<html><body>Search results for: " + q + "</body></html>"))
	}))
	defer mockServer.Close()

	rule := assessment.NewInputValidationRule()
	target := &models.Target{ID: "tgt-1", Domain: "example.com"}

	traffic := []models.TrafficRecord{
		{
			URL:                mockServer.URL + "/search?q=test",
			Method:             "GET",
			ResponseStatusCode: 200,
		},
	}

	ctx := context.Background()
	findings, err := rule.Evaluate(ctx, target, traffic, mockServer.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 reflection finding, got %d", len(findings))
	}

	finding := findings[0]
	if !strings.Contains(finding.Name, "Unescaped Parameter Reflection") {
		t.Errorf("unexpected finding name: %s", finding.Name)
	}
	if !strings.Contains(finding.EvidenceResponse, "<canary>") {
		t.Errorf("expected evidence response to contain reflected canary")
	}
}

func TestUploadSecurityRule(t *testing.T) {
	rule := assessment.NewUploadSecurityRule()
	target := &models.Target{ID: "tgt-1", Domain: "example.com"}

	traffic := []models.TrafficRecord{
		{
			URL:                "https://example.com/api/upload",
			Method:             "POST",
			Category:           "upload",
			ContentType:        "multipart/form-data; boundary=---boundary",
			RequestBody:        "---boundary\r\nContent-Disposition: form-data; name=\"file\"; filename=\"webshell.php\"\r\n\r\n<?php phpinfo(); ?>\r\n---boundary--",
			ResponseStatusCode: 200,
		},
	}

	ctx := context.Background()
	findings, err := rule.Evaluate(ctx, target, traffic, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("expected 1 upload finding, got %d", len(findings))
	}
	if !strings.Contains(findings[0].Name, "Dangerous File Extension Accepted") {
		t.Errorf("unexpected finding name: %s", findings[0].Name)
	}
}

func TestAssessmentEngine(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "assessment_test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()
	_ = database.MigrateSchema(db)

	targetRepo := database.NewTargetRepo(db)
	scanRepo := database.NewScanRepo(db)
	findingRepo := database.NewFindingRepo(db)
	target, _ := targetRepo.Create("Engine Target", "test.example.com", "")

	_ = scanRepo.Create(&models.Scan{
		ID:        "scan-test-1",
		TargetID:  target.ID,
		ToolName:  "assessment",
		Status:    models.StatusCompleted,
		Command:   "assessment",
		StartedAt: time.Now().UTC(),
	})

	engine := assessment.NewAssessmentEngine(findingRepo, nil)
	if len(engine.Rules()) < 4 {
		t.Errorf("expected at least 4 registered rules, got %d", len(engine.Rules()))
	}

	traffic := []models.TrafficRecord{
		{
			URL:                "https://test.example.com/login?password=admin",
			Method:             "GET",
			ResponseStatusCode: 200,
		},
	}

	ctx := context.Background()
	findings, err := engine.Run(ctx, target, traffic, "scan-test-1")
	if err != nil {
		t.Fatalf("engine.Run failed: %v", err)
	}

	if len(findings) == 0 {
		t.Fatal("expected at least 1 finding from engine run")
	}

	// Verify finding saved to database
	dbFindings, total, err := findingRepo.ListByTarget(target.ID, "", "", 10, 0)
	if err != nil {
		t.Fatalf("failed to list findings from db: %v", err)
	}
	if total == 0 || len(dbFindings) == 0 {
		t.Errorf("expected finding to be persisted to SQLite database, got total %d", total)
	}
}

func TestBuildCurl(t *testing.T) {
	req, _ := http.NewRequest("POST", "https://example.com/api/v1/data", io.NopCloser(strings.NewReader(`{"key":"value"}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer testtoken")

	curl := assessment.BuildCurl(req, `{"key":"value"}`)
	if !strings.Contains(curl, "curl -i -s -k") {
		t.Errorf("missing base curl command: %s", curl)
	}
	if !strings.Contains(curl, "-X POST") {
		t.Errorf("missing POST method: %s", curl)
	}
	if !strings.Contains(curl, "-H \"Authorization: Bearer testtoken\"") {
		t.Errorf("missing Authorization header in curl: %s", curl)
	}
	if !strings.Contains(curl, `--data-raw '{"key":"value"}'`) {
		t.Errorf("missing data-raw in curl: %s", curl)
	}
}
