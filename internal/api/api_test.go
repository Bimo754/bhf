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
	"github.com/diamond/bhf/internal/assessment"
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

func TestTrafficHandler(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "test_traffic_api.db")
	db, err := database.Open(tempDB)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := database.MigrateSchema(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	targetRepo := database.NewTargetRepo(db)
	trafficRepo := database.NewTrafficRepo(db)

	target, err := targetRepo.Create("Traffic API Org", "api.example.com", "")
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	rec1 := &models.TrafficRecord{
		ID:                 "tr-1",
		TargetID:           target.ID,
		Method:             "POST",
		URL:                "https://api.example.com/login",
		Path:               "/login",
		Host:               "api.example.com",
		ResponseStatusCode: 200,
		Category:           "auth",
		CreatedAt:          time.Now().UTC(),
	}
	if err := trafficRepo.Insert(rec1); err != nil {
		t.Fatalf("failed to insert traffic record: %v", err)
	}

	trafficHandler := api.NewTrafficHandler(trafficRepo)
	r := chi.NewRouter()
	r.Get("/api/traffic/target/{targetID}", trafficHandler.ListByTarget)
	r.Get("/api/traffic/target/{targetID}/stats", trafficHandler.GetStats)
	r.Get("/api/traffic/{id}", trafficHandler.GetByID)

	// Test 1: List by target
	req1 := httptest.NewRequest(http.MethodGet, "/api/traffic/target/"+target.ID+"?category=auth", nil)
	recOut1 := httptest.NewRecorder()
	r.ServeHTTP(recOut1, req1)
	if recOut1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recOut1.Code)
	}

	var listEnvelope struct {
		Success bool                    `json:"success"`
		Data    api.TrafficListResponse `json:"data"`
	}
	if err := json.Unmarshal(recOut1.Body.Bytes(), &listEnvelope); err != nil {
		t.Fatalf("failed to parse list resp: %v", err)
	}
	if listEnvelope.Data.Total != 1 || len(listEnvelope.Data.Traffic) != 1 {
		t.Errorf("expected 1 traffic item, got total %d, len %d", listEnvelope.Data.Total, len(listEnvelope.Data.Traffic))
	}

	// Test 2: Get by ID
	req2 := httptest.NewRequest(http.MethodGet, "/api/traffic/tr-1", nil)
	recOut2 := httptest.NewRecorder()
	r.ServeHTTP(recOut2, req2)
	if recOut2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recOut2.Code)
	}

	var itemEnvelope struct {
		Success bool                 `json:"success"`
		Data    models.TrafficRecord `json:"data"`
	}
	if err := json.Unmarshal(recOut2.Body.Bytes(), &itemEnvelope); err != nil {
		t.Fatalf("failed to parse item resp: %v", err)
	}
	if itemEnvelope.Data.ID != "tr-1" || itemEnvelope.Data.Path != "/login" {
		t.Errorf("unexpected item: %+v", itemEnvelope.Data)
	}

	// Test 3: Get Stats
	req3 := httptest.NewRequest(http.MethodGet, "/api/traffic/target/"+target.ID+"/stats", nil)
	recOut3 := httptest.NewRecorder()
	r.ServeHTTP(recOut3, req3)
	if recOut3.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recOut3.Code)
	}

	// Test 4: Not Found
	req4 := httptest.NewRequest(http.MethodGet, "/api/traffic/nonexistent", nil)
	recOut4 := httptest.NewRecorder()
	r.ServeHTTP(recOut4, req4)
	if recOut4.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing traffic record, got %d", recOut4.Code)
	}
}

func TestAssessmentAndNetworkHandlers(t *testing.T) {
	tempDB := filepath.Join(t.TempDir(), "test_api_ext.db")
	db, err := database.Open(tempDB)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	if err := database.MigrateSchema(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	targetRepo := database.NewTargetRepo(db)
	scanRepo := database.NewScanRepo(db)
	assetRepo := database.NewAssetRepo(db)
	trafficRepo := database.NewTrafficRepo(db)
	findingRepo := database.NewFindingRepo(db)
	hostRepo := database.NewHostRepo(db)
	serviceRepo := database.NewServiceRepo(db)

	target, _ := targetRepo.Create("Network Scope", "corp.internal", "AD network")
	engine := assessment.NewAssessmentEngine(findingRepo, nil)

	// Insert sample traffic with query credential leak
	_ = trafficRepo.Insert(&models.TrafficRecord{
		ID:                 "tr-auth-1",
		TargetID:           target.ID,
		Method:             "GET",
		URL:                "https://corp.internal/auth?api_key=secret123",
		Path:               "/auth",
		Host:               "corp.internal",
		ResponseStatusCode: 200,
		Category:           "auth",
	})

	// 1. Test Assessment Handler
	assessmentHandler := api.NewAssessmentHandler(engine, trafficRepo, targetRepo, scanRepo)
	r := chi.NewRouter()
	r.Get("/api/assessment/rules", assessmentHandler.ListRules)
	r.Post("/api/assessment/target/{targetID}/run", assessmentHandler.RunTargetAssessment)

	// List rules
	req1 := httptest.NewRequest(http.MethodGet, "/api/assessment/rules", nil)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200 from ListRules, got %d", rec1.Code)
	}

	// Run assessment
	req2 := httptest.NewRequest(http.MethodPost, "/api/assessment/target/"+target.ID+"/run", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 from RunTargetAssessment, got %d (%s)", rec2.Code, rec2.Body.String())
	}

	var runResp struct {
		Success bool            `json:"success"`
		Data    api.RunResponse `json:"data"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &runResp); err != nil {
		t.Fatalf("failed to parse run response: %v", err)
	}
	if runResp.Data.TrafficAnalyzed != 1 || runResp.Data.FindingsCount == 0 {
		t.Errorf("expected findings from assessment, got %+v", runResp.Data)
	}

	// 2. Test Network Handler
	_ = hostRepo.Upsert(&models.Host{
		TargetID: target.ID,
		IP:       "10.0.0.1",
		Status:   "up",
	})

	networkHandler := api.NewNetworkHandler(hostRepo, serviceRepo, assetRepo, findingRepo, targetRepo, scanRepo)
	r.Get("/api/network/target/{targetID}/hosts", networkHandler.ListHosts)
	r.Get("/api/network/target/{targetID}/services", networkHandler.ListServices)
	r.Post("/api/network/target/{targetID}/ad-audit", networkHandler.AuditActiveDirectory)

	req3 := httptest.NewRequest(http.MethodGet, "/api/network/target/"+target.ID+"/hosts", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200 from ListHosts, got %d", rec3.Code)
	}

	reqAD := httptest.NewRequest(http.MethodPost, "/api/network/target/"+target.ID+"/ad-audit", nil)
	recAD := httptest.NewRecorder()
	r.ServeHTTP(recAD, reqAD)
	if recAD.Code != http.StatusOK {
		t.Fatalf("expected 200 from AuditActiveDirectory, got %d", recAD.Code)
	}

	// 3. Test AppIntel Handler
	appIntelHandler := api.NewAppIntelHandler(trafficRepo, targetRepo, scanRepo)
	r.Get("/api/appintel/proxy/status", appIntelHandler.GetProxyStatus)

	req4 := httptest.NewRequest(http.MethodGet, "/api/appintel/proxy/status", nil)
	rec4 := httptest.NewRecorder()
	r.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Fatalf("expected 200 from GetProxyStatus, got %d", rec4.Code)
	}
}
