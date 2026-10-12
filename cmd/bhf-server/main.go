package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/diamond/bhf/internal/api"
	"github.com/diamond/bhf/internal/config"
	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/engine"
	"github.com/diamond/bhf/internal/tools"
	"github.com/diamond/bhf/web"
)

func main() {
	cfg := config.Load()
	log.Printf("[BHF] Starting Bug Hunting Framework on %s:%d...", cfg.Host, cfg.Port)

	// Initialize SQLite Database with WAL mode
	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("[BHF] Fatal database connection error: %v", err)
	}
	defer db.Close()

	if err := database.MigrateSchema(db); err != nil {
		log.Fatalf("[BHF] Fatal database migration error: %v", err)
	}
	log.Printf("[BHF] SQLite database ready at %s", cfg.DatabasePath)

	// Repositories
	targetRepo := database.NewTargetRepo(db)
	scanRepo := database.NewScanRepo(db)
	assetRepo := database.NewAssetRepo(db)
	logRepo := database.NewLogRepo(db)

	// Engine & Streaming
	hub := engine.NewHub()
	runner := engine.NewRunner(logRepo, hub)
	dispatcher := engine.NewDispatcher(runner, scanRepo, hub)

	// Tools Registry
	registry := tools.NewRegistry()
	registry.Register(tools.NewSubfinderTool())
	registry.Register(tools.NewHttpxTool())
	registry.Register(tools.NewAssetfinderTool())
	registry.Register(tools.NewGauTool())
	registry.Register(tools.NewKatanaTool())
	registry.Register(tools.NewNucleiTool())
	registry.Register(tools.NewFfufTool())

	log.Printf("[BHF] Registered %d security tools", len(registry.List()))

	// Handlers
	targetHandler := api.NewTargetHandler(targetRepo)
	scanHandler := api.NewScanHandler(scanRepo, targetRepo, assetRepo, logRepo, dispatcher, runner, registry, hub)
	assetHandler := api.NewAssetHandler(assetRepo)
	toolHandler := api.NewToolHandler(registry)
	wsHandler := api.NewWSHandler(hub)

	staticFS, err := web.GetFileSystem()
	if err != nil {
		log.Printf("[BHF] Warning: Could not load embedded static files: %v", err)
	}

	router := api.NewRouter(api.RouterConfig{
		TargetHandler: targetHandler,
		ScanHandler:   scanHandler,
		AssetHandler:  assetHandler,
		ToolHandler:   toolHandler,
		WSHandler:     wsHandler,
		StaticFS:      staticFS,
	})

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Server runner
	go func() {
		log.Printf("[BHF] Server listening on http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[BHF] Listen and serve error: %v", err)
		}
	}()

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[BHF] Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[BHF] Server forced to shutdown: %v", err)
	}

	log.Println("[BHF] Server stopped cleanly.")
}
