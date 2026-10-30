package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// RouterConfig holds all handler dependencies to initialize the HTTP routing table.
type RouterConfig struct {
	TargetHandler  *TargetHandler
	ScanHandler    *ScanHandler
	AssetHandler    *AssetHandler
	EndpointHandler *EndpointHandler
	FindingHandler  *FindingHandler
	ToolHandler     *ToolHandler
	WSHandler       *WSHandler
	MCPServer       http.Handler
	StaticFS        http.FileSystem
	DataDir         string
}

// NewRouter constructs a configured chi HTTP router.
func NewRouter(cfg RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Standard middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Cross-Origin Resource Sharing
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API Endpoints
	r.Route("/api", func(api chi.Router) {
		// Target management
		api.Route("/targets", func(t chi.Router) {
			t.Get("/", cfg.TargetHandler.List)
			t.Post("/", cfg.TargetHandler.Create)
			t.Get("/{id}", cfg.TargetHandler.GetByID)
			t.Delete("/{id}", cfg.TargetHandler.Delete)
		})

		// Tool executions & scans
		api.Route("/scans", func(s chi.Router) {
			s.Post("/start", cfg.ScanHandler.Start)
			s.Post("/{id}/stop", cfg.ScanHandler.Stop)
			s.Get("/{id}/logs", cfg.ScanHandler.GetLogs)
			s.Get("/target/{targetID}", cfg.ScanHandler.ListByTarget)
		})

		// Discovered assets
		api.Route("/assets", func(a chi.Router) {
			a.Get("/target/{targetID}", cfg.AssetHandler.ListByTarget)
		})

		// Discovered endpoints (FEAT-017)
		if cfg.EndpointHandler != nil {
			api.Route("/endpoints", func(e chi.Router) {
				e.Get("/target/{targetID}", cfg.EndpointHandler.ListByTarget)
			})
		}

		// Security findings
		api.Route("/findings", func(f chi.Router) {
			f.Get("/target/{targetID}", cfg.FindingHandler.ListByTarget)
			f.Delete("/{id}", cfg.FindingHandler.Delete)
		})

		// Supported tools catalog
		api.Route("/tools", func(tl chi.Router) {
			tl.Get("/", cfg.ToolHandler.List)
		})

		// Model Context Protocol (MCP) AI integration
		if cfg.MCPServer != nil {
			api.Post("/mcp", cfg.MCPServer.ServeHTTP)
		}

		// WebSocket streaming
		api.Get("/ws/scans/{id}/logs", cfg.WSHandler.StreamLogs)

		// Static screenshots file server with intelligent path & hash fallback
		if cfg.DataDir != "" {
			screenshotsDir := filepath.Join(cfg.DataDir, "screenshots")
			_ = os.MkdirAll(screenshotsDir, 0755)
			_ = filepath.WalkDir(screenshotsDir, func(path string, d os.DirEntry, err error) error {
				if err == nil {
					if d.IsDir() {
						_ = os.Chmod(path, 0755)
					} else {
						_ = os.Chmod(path, 0644)
					}
				}
				return nil
			})
			api.Handle("/screenshots/*", screenshotHandler(screenshotsDir))
		}
	})

	// Static frontend assets (if provided)
	if cfg.StaticFS != nil {
		fileServer := http.FileServer(cfg.StaticFS)
		r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			path := req.URL.Path
			// Try opening the file in staticFS
			f, err := cfg.StaticFS.Open(path)
			if err != nil {
				// Fallback to index.html for SPA client-side routing
				req.URL.Path = "/"
			} else {
				_ = f.Close()
			}
			fileServer.ServeHTTP(w, req)
		}))
	}

	return r
}

// screenshotHandler provides a secure, fallback-aware file server for visual screenshots.
// It serves exact relative paths if they exist, or falls back to resolving by filename within
// the target's screenshot tree (supporting both flat and nested httpx -srd structures).
func screenshotHandler(screenshotsDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqPath := strings.TrimPrefix(r.URL.Path, "/api/screenshots/")
		reqPath = strings.TrimPrefix(reqPath, "/screenshots/")
		cleanPath := filepath.Clean(reqPath)

		// Directory traversal guard
		if cleanPath == "." || strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, "../") || strings.Contains(cleanPath, `..\`) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		fullPath := filepath.Join(screenshotsDir, cleanPath)

		// 1. Direct match exists and is a file
		if fi, err := os.Stat(fullPath); err == nil && !fi.IsDir() {
			w.Header().Set("Cache-Control", "public, max-age=86400")
			http.ServeFile(w, r, fullPath)
			return
		}

		// 2. Fallback: search within target directory by filename
		parts := strings.Split(cleanPath, string(filepath.Separator))
		if len(parts) >= 2 {
			targetID := parts[0]
			targetDir := filepath.Join(screenshotsDir, targetID)
			fileName := filepath.Base(cleanPath)

			var foundPath string
			if fi, err := os.Stat(targetDir); err == nil && fi.IsDir() {
				_ = filepath.WalkDir(targetDir, func(p string, d os.DirEntry, err error) error {
					if err != nil {
						return nil
					}
					if !d.IsDir() && d.Name() == fileName {
						foundPath = p
						return filepath.SkipAll
					}
					return nil
				})
			}

			if foundPath != "" {
				w.Header().Set("Cache-Control", "public, max-age=86400")
				http.ServeFile(w, r, foundPath)
				return
			}
		}

		http.NotFound(w, r)
	}
}

