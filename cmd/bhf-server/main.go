package main

import (
	"context"
	"flag"
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
	"github.com/diamond/bhf/internal/mcp"
	"github.com/diamond/bhf/internal/tools"
	"github.com/diamond/bhf/web"
)

func main() {
	mcpMode := flag.Bool("mcp", false, "Run in Model Context Protocol (MCP) stdio mode for AI agents")
	flag.Parse()

	cfg := config.Load()
	if !*mcpMode {
		log.Printf("[BHF] Starting Bug Hunting Framework on %s:%d...", cfg.Host, cfg.Port)
	}

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
	findingRepo := database.NewFindingRepo(db)
	endpointRepo := database.NewEndpointRepo(db)
	logRepo := database.NewLogRepo(db)

	// Engine & Streaming
	hub := engine.NewHub()
	runner := engine.NewRunner(logRepo, hub)
	dispatcher := engine.NewDispatcher(runner, scanRepo, hub)

	// Tools Registry
	registry := tools.NewRegistry()
	registry.Register(tools.NewSubfinderTool())
	registry.Register(tools.NewAssetfinderTool())
	registry.Register(tools.NewTheHarvesterTool())
	registry.Register(tools.NewDNSxTool())
	registry.Register(tools.NewNaabuTool())
	registry.Register(tools.NewHttpxTool())
	registry.Register(tools.NewTlsxTool())
	registry.Register(tools.NewWafw00fTool())
	registry.Register(tools.NewGowitnessTool())
	registry.Register(tools.NewKatanaTool())
	registry.Register(tools.NewGauTool())
	registry.Register(tools.NewUroTool())
	registry.Register(tools.NewJsluiceTool())
	registry.Register(tools.NewFfufTool())
	registry.Register(tools.NewArjunTool())
	registry.Register(tools.NewNomore403Tool())
	registry.Register(tools.NewSubzyTool())
	registry.Register(tools.NewNucleiTool())
	registry.Register(tools.NewCorsyTool())
	registry.Register(tools.NewS3ScannerTool())

	log.Printf("[BHF] Registered %d security tools", len(registry.List()))

	// Handlers
	targetHandler := api.NewTargetHandler(targetRepo)
	scanHandler := api.NewScanHandler(scanRepo, targetRepo, assetRepo, findingRepo, endpointRepo, logRepo, dispatcher, runner, registry, hub)
	assetHandler := api.NewAssetHandler(assetRepo)
	endpointHandler := api.NewEndpointHandler(endpointRepo)
	findingHandler := api.NewFindingHandler(findingRepo)
	toolHandler := api.NewToolHandler(registry)
	wsHandler := api.NewWSHandler(hub)

	staticFS, err := web.GetFileSystem()
	if err != nil {
		log.Printf("[BHF] Warning: Could not load embedded static files: %v", err)
	}

	// MCP Server (Model Context Protocol for AI Agent Interoperability)
	mcpServer := mcp.NewServer(targetRepo, assetRepo, findingRepo, registry)
	if *mcpMode {
		log.Println("[BHF] Starting in MCP stdio mode for AI agents...")
		if err := mcpServer.RunStdio(context.Background()); err != nil {
			log.Fatalf("[BHF] MCP stdio error: %v", err)
		}
		return
	}

	router := api.NewRouter(api.RouterConfig{
		TargetHandler:   targetHandler,
		ScanHandler:     scanHandler,
		AssetHandler:    assetHandler,
		EndpointHandler: endpointHandler,
		FindingHandler:  findingHandler,
		ToolHandler:     toolHandler,
		WSHandler:       wsHandler,
		MCPServer:       mcpServer,
		StaticFS:        staticFS,
		DataDir:         cfg.DataDir,
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
