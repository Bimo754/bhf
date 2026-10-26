package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// RouterConfig holds all handler dependencies to initialize the HTTP routing table.
type RouterConfig struct {
	TargetHandler  *TargetHandler
	ScanHandler    *ScanHandler
	AssetHandler   *AssetHandler
	FindingHandler *FindingHandler
	ToolHandler    *ToolHandler
	WSHandler      *WSHandler
	MCPServer      http.Handler
	StaticFS       http.FileSystem
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
