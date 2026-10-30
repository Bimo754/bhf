package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/diamond/bhf/internal/api"
	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/go-chi/chi/v5"
)

func setupTestDB(t *testing.T) (*database.TargetRepo, *database.ScanRepo, *database.AssetRepo, *database.EndpointRepo, *models.Target, *models.Scan) {
	tempDB := filepath.Join(t.TempDir(), "test_api.db")
	db, err := database.Open(tempDB)
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := database.MigrateSchema(db); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	targetRepo := database.NewTargetRepo(db)
	scanRepo := database.NewScanRepo(db)
	assetRepo := database.NewAssetRepo(db)
	endpointRepo := database.NewEndpointRepo(db)

	target, err := targetRepo.Create("Test Org", "example.com", "Test scope")
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	scan := &models.Scan{
		ID:        "test-scan-1",
		TargetID:  target.ID,
		ToolName:  "httpx",
		Status:    models.StatusCompleted,
		Command:   "httpx",
		StartedAt: time.Now().UTC(),
	}
	if err := scanRepo.Create(scan); err != nil {
		t.Fatalf("failed to create scan: %v", err)
	}

	return targetRepo, scanRepo, assetRepo, endpointRepo, target, scan
}

func TestAssetHandlerPaginationAndFilter(t *testing.T) {
	_, _, assetRepo, _, target, scan := setupTestDB(t)
	now := time.Now().UTC()

	// Insert test assets
	if err := assetRepo.Upsert(&models.Asset{
		ID:            "asset-1",
		TargetID:      target.ID,
		ScanID:        scan.ID,
		Host:          "sub1.example.com",
		StatusCode:    0,
		PriorityScore: 10,
		CreatedAt:     now,
	}); err != nil {
		t.Fatalf("failed to upsert asset-1: %v", err)
	}

	if err := assetRepo.Upsert(&models.Asset{
		ID:            "asset-2",
		TargetID:      target.ID,
		ScanID:        scan.ID,
		Host:          "admin.example.com",
		URL:           "https://admin.example.com",
		StatusCode:    200,
		PriorityScore: 50,
		CreatedAt:     now,
	}); err != nil {
		t.Fatalf("failed to upsert asset-2: %v", err)
	}

	handler := api.NewAssetHandler(assetRepo)
	r := chi.NewRouter()
	r.Get("/api/assets/target/{targetID}", handler.ListByTarget)

	// Test 1: Query with filter=live
	req := httptest.NewRequest(http.MethodGet, "/api/assets/target/"+target.ID+"?filter=live", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var envelope struct {
		Success bool                  `json:"success"`
		Data    api.AssetListResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if envelope.Data.Total != 1 {
		t.Errorf("expected 1 live asset, got total %d", envelope.Data.Total)
	}
	if len(envelope.Data.Assets) != 1 || envelope.Data.Assets[0].Host != "admin.example.com" {
		t.Errorf("expected admin.example.com asset, got %+v", envelope.Data.Assets)
	}
}

func TestEndpointHandlerListAndSitemap(t *testing.T) {
	_, _, _, endpointRepo, target, _ := setupTestDB(t)
	now := time.Now().UTC()

	if err := endpointRepo.Upsert(&models.Endpoint{
		ID:         "ep-1",
		TargetID:   target.ID,
		Host:       "api.example.com",
		URL:        "https://api.example.com/v1/users",
		Path:       "/v1/users",
		Method:     "GET",
		StatusCode: 200,
		CreatedAt:  now,
	}); err != nil {
		t.Fatalf("failed to upsert ep-1: %v", err)
	}
	if err := endpointRepo.Upsert(&models.Endpoint{
		ID:         "ep-2",
		TargetID:   target.ID,
		Host:       "api.example.com",
		URL:        "https://api.example.com/v1/auth",
		Path:       "/v1/auth",
		Method:     "POST",
		StatusCode: 401,
		CreatedAt:  now,
	}); err != nil {
		t.Fatalf("failed to upsert ep-2: %v", err)
	}

	handler := api.NewEndpointHandler(endpointRepo)
	r := chi.NewRouter()
	r.Get("/api/endpoints/target/{targetID}", handler.ListByTarget)

	// Test 1: Query with all=true for sitemap tree
	req := httptest.NewRequest(http.MethodGet, "/api/endpoints/target/"+target.ID+"?all=true", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var epEnvelope struct {
		Success bool                     `json:"success"`
		Data    api.EndpointListResponse `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&epEnvelope); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if epEnvelope.Data.Total != 2 || len(epEnvelope.Data.Endpoints) != 2 {
		t.Errorf("expected 2 endpoints in sitemap response, got %d", epEnvelope.Data.Total)
	}
}

func TestScreenshotHandler(t *testing.T) {
	dataDir := t.TempDir()
	targetID := "test-target-uuid"
	targetScreenshots := filepath.Join(dataDir, "screenshots", targetID)
	nestedDir := filepath.Join(targetScreenshots, "screenshot", "sub.example.com")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}

	testPNG := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG magic bytes
	testFile := filepath.Join(nestedDir, "abc123hash.png")
	if err := os.WriteFile(testFile, testPNG, 0644); err != nil {
		t.Fatalf("failed to write test png: %v", err)
	}

	r := api.NewRouter(api.RouterConfig{
		DataDir: dataDir,
	})

	// Test 1: Exact nested path returns 200 OK
	req1 := httptest.NewRequest(http.MethodGet, "/api/screenshots/"+targetID+"/screenshot/sub.example.com/abc123hash.png", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Errorf("expected status 200 for exact path, got %d", rec1.Code)
	}

	// Test 2: Fallback direct filename without subdirs returns 200 OK via hash search
	req2 := httptest.NewRequest(http.MethodGet, "/api/screenshots/"+targetID+"/abc123hash.png", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Errorf("expected status 200 for fallback search, got %d", rec2.Code)
	}

	// Test 3: Path traversal attempt returns 403 Forbidden
	req3 := httptest.NewRequest(http.MethodGet, "/api/screenshots/../../etc/passwd", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusForbidden && rec3.Code != http.StatusNotFound {
		t.Errorf("expected 403 or 404 for traversal attempt, got %d", rec3.Code)
	}

	// Test 4: Nonexistent file returns 404 Not Found
	req4 := httptest.NewRequest(http.MethodGet, "/api/screenshots/"+targetID+"/nonexistent.png", nil)
	rec4 := httptest.NewRecorder()
	r.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing file, got %d", rec4.Code)
	}
}
