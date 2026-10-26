package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/tools"
)

func TestMCPServerProtocol(t *testing.T) {
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory db: %v", err)
	}
	defer db.Close()

	if err := database.MigrateSchema(db); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	targetRepo := database.NewTargetRepo(db)
	assetRepo := database.NewAssetRepo(db)
	findingRepo := database.NewFindingRepo(db)
	registry := tools.NewRegistry()
	registry.Register(tools.NewSubfinderTool())

	target, err := targetRepo.Create("Test Org", "test.com", "Bug bounty scope")
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	srv := NewServer(targetRepo, assetRepo, findingRepo, registry)
	ctx := context.Background()

	// 1. Test initialize
	initReq := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	resp := srv.HandleMessage(ctx, initReq)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected successful initialize response, got: %+v", resp)
	}

	// 2. Test tools/list
	listReq := []byte(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	resp = srv.HandleMessage(ctx, listReq)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected successful tools/list response, got: %+v", resp)
	}

	// 3. Test tools/call list_targets
	callReq := []byte(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_targets","arguments":{}}}`)
	resp = srv.HandleMessage(ctx, callReq)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected successful tools/call response, got: %+v", resp)
	}

	// 4. Test tools/call get_target_stats
	statsArgs, _ := json.Marshal(map[string]any{
		"name": "get_target_stats",
		"arguments": map[string]any{
			"target_id": target.ID,
		},
	})
	callStatsReq := []byte(`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":` + string(statsArgs) + `}`)
	resp = srv.HandleMessage(ctx, callStatsReq)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected successful get_target_stats response, got: %+v", resp)
	}
}
