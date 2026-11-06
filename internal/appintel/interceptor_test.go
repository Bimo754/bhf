package appintel_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/diamond/bhf/internal/appintel"
	"github.com/diamond/bhf/internal/database"
)

func TestRecordingTransport(t *testing.T) {
	// Set up isolated SQLite database
	dbPath := filepath.Join(t.TempDir(), "interceptor_test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := database.MigrateSchema(db); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	targetRepo := database.NewTargetRepo(db)
	trafficRepo := database.NewTrafficRepo(db)

	target, err := targetRepo.Create("AppIntel Target", "example.com", "")
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	// Mock target server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.URL.Path == "/api/login" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Set-Cookie", "sessionid=xyz123; Path=/")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"success":true,"user":"alice"}`))
			return
		}
		if r.URL.Path == "/echo" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write(body)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	serverHost := serverURL.Hostname()

	cfg := appintel.InterceptorConfig{
		TargetID:     target.ID,
		TargetDomain: serverHost, // Scope matches mock server
		TrafficRepo:  trafficRepo,
		MaxBodyBytes: 1024 * 1024,
	}

	client := &http.Client{
		Transport: appintel.NewRecordingTransport(cfg, http.DefaultTransport),
	}

	// 1. Send Login Request (In-Scope)
	loginReqBody := `{"username":"alice","password":"secretpassword"}`
	loginReq, _ := http.NewRequest("POST", server.URL+"/api/login", strings.NewReader(loginReqBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := client.Do(loginReq)
	if err != nil {
		t.Fatalf("failed to do login request: %v", err)
	}
	defer loginResp.Body.Close()

	resBody, _ := io.ReadAll(loginResp.Body)
	if !strings.Contains(string(resBody), "alice") {
		t.Errorf("unexpected response body: %s", string(resBody))
	}

	// Verify recorded in SQLite
	records, total, err := trafficRepo.ListByTarget(target.ID, "", "", "", 10, 0)
	if err != nil {
		t.Fatalf("failed to query recorded traffic: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 recorded transaction, got %d", total)
	}

	rec := records[0]
	if rec.Method != "POST" {
		t.Errorf("expected POST method, got %s", rec.Method)
	}
	if rec.Path != "/api/login" {
		t.Errorf("expected /api/login path, got %s", rec.Path)
	}
	if rec.Category != "auth" {
		t.Errorf("expected auth category, got %s", rec.Category)
	}
	if rec.ResponseStatusCode != 200 {
		t.Errorf("expected status 200, got %d", rec.ResponseStatusCode)
	}
	if !strings.Contains(rec.RequestBody, "secretpassword") {
		t.Errorf("expected request body to be recorded, got %s", rec.RequestBody)
	}

	// 2. Out-of-Scope Scope Guard Test
	outOfScopeCfg := appintel.InterceptorConfig{
		TargetID:     target.ID,
		TargetDomain: "different-domain.com", // Out of scope for serverHost
		TrafficRepo:  trafficRepo,
	}
	outOfScopeClient := &http.Client{
		Transport: appintel.NewRecordingTransport(outOfScopeCfg, http.DefaultTransport),
	}

	outReq, _ := http.NewRequest("GET", server.URL+"/echo", bytes.NewBufferString("hello"))
	outResp, err := outOfScopeClient.Do(outReq)
	if err != nil {
		t.Fatalf("out of scope request failed: %v", err)
	}
	defer outResp.Body.Close()

	// Should not have incremented record count
	recordsAfter, totalAfter, _ := trafficRepo.ListByTarget(target.ID, "", "", "", 10, 0)
	if totalAfter != 1 || len(recordsAfter) != 1 {
		t.Errorf("out-of-scope request was recorded in database! Expected total 1, got %d", totalAfter)
	}
}

func TestProxyServer(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "proxy_test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()
	_ = database.MigrateSchema(db)

	targetRepo := database.NewTargetRepo(db)
	trafficRepo := database.NewTrafficRepo(db)

	target, _ := targetRepo.Create("Proxy Target", "127.0.0.1", "")

	// Upstream test server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "bhf-proxy-success")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("proxied response"))
	}))
	defer upstream.Close()

	cfg := appintel.InterceptorConfig{
		TargetID:     target.ID,
		TargetDomain: "127.0.0.1",
		TrafficRepo:  trafficRepo,
	}

	proxy := appintel.NewProxyServer(cfg)
	proxyAddr, err := proxy.Start("127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start proxy server: %v", err)
	}
	defer proxy.Stop()

	proxyURL, _ := url.Parse("http://" + proxyAddr)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	resp, err := client.Get(upstream.URL)
	if err != nil {
		t.Fatalf("client request through proxy failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "proxied response" {
		t.Errorf("unexpected body through proxy: %s", string(body))
	}
	if resp.Header.Get("X-Custom-Header") != "bhf-proxy-success" {
		t.Errorf("missing custom header from upstream")
	}

	// Verify proxy persisted transaction to database
	count, err := trafficRepo.CountByTarget(target.ID)
	if err != nil {
		t.Fatalf("failed to count traffic records: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 traffic record via proxy, got %d", count)
	}
}
