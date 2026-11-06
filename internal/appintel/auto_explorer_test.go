package appintel_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/diamond/bhf/internal/appintel"
	"github.com/diamond/bhf/internal/database"
)

func TestAutoExplorerHTTPSpider(t *testing.T) {
	// Create mock target web application with multiple pages and forms
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<head><title>Test App</title></head>
			<body>
				<h1>Welcome to Target App</h1>
				<a href="/login">Login Page</a>
				<a href="/search">Search Products</a>
				<a href="https://external-out-of-scope.com/ignored">Out of scope</a>
			</body>
			</html>
		`))
	})

	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"logged_in"}`))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<body>
				<form action="/login" method="POST">
					<input type="text" name="username" />
					<input type="password" name="password" />
					<button type="submit">Submit</button>
				</form>
				<a href="/dashboard">Dashboard</a>
			</body>
			</html>
		`))
	})

	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<!DOCTYPE html>
			<html>
			<body>
				<form action="/search" method="GET">
					<input type="text" name="q" />
				</form>
				<a href="/">Home</a>
			</body>
			</html>
		`))
	})

	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<div>Admin Dashboard</div>`))
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Setup in-memory SQLite database
	tempDB := filepath.Join(t.TempDir(), "test_explorer.db")
	db, err := database.Open(tempDB)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := database.MigrateSchema(db); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	targetRepo := database.NewTargetRepo(db)
	trafficRepo := database.NewTrafficRepo(db)

	// Clean host for target domain
	serverURL := ts.URL
	targetHost := strings.TrimPrefix(serverURL, "http://")

	target, err := targetRepo.Create("Autonomous Target", targetHost, "Test site")
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	cfg := appintel.AutoExplorerConfig{
		Target:       target,
		TrafficRepo:  trafficRepo,
		ScanID:       "test-scan-auto",
		StartURL:     ts.URL,
		MaxPages:     10,
		MaxDepth:     3,
		PageTimeout:  2 * time.Second,
		TotalTimeout: 10 * time.Second,
		PreferChrome: false, // Pure-Go HTTP spider mode for fast CI test
	}

	var updates []appintel.ExplorerProgress
	explorer := appintel.NewAutoExplorer(cfg, func(p appintel.ExplorerProgress) {
		updates = append(updates, p)
	})

	ctx := context.Background()
	progress, err := explorer.Run(ctx)
	if err != nil {
		t.Fatalf("explorer run failed: %v", err)
	}

	if progress.PagesVisited < 3 {
		t.Errorf("expected at least 3 pages visited, got %d", progress.PagesVisited)
	}

	// Verify traffic was recorded into database
	traffic, total, err := trafficRepo.ListByTarget(target.ID, "", "", "", 50, 0)
	if err != nil {
		t.Fatalf("failed to list traffic: %v", err)
	}

	if total == 0 || len(traffic) == 0 {
		t.Fatalf("expected recorded traffic transactions, got %d", total)
	}

	// Verify categories were inferred (e.g. auth for /login)
	foundAuth := false
	for _, rec := range traffic {
		if strings.Contains(rec.Path, "/login") && rec.Category == "auth" {
			foundAuth = true
			break
		}
	}
	if !foundAuth {
		t.Errorf("expected /login request to be categorized as 'auth'")
	}
}
