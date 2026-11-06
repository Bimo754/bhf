package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/diamond/bhf/internal/assessment"
	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
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

	// 5. Test tools/call query_traffic and get_parameter_graph
	trafficRepo := database.NewTrafficRepo(db)
	srv.SetTrafficRepo(trafficRepo)

	trafficArgs, _ := json.Marshal(map[string]any{
		"name": "query_traffic",
		"arguments": map[string]any{
			"target_id": target.ID,
		},
	})
	callTrafficReq := []byte(`{"jsonrpc":"2.0","id":5,"method":"tools/call","params":` + string(trafficArgs) + `}`)
	resp = srv.HandleMessage(ctx, callTrafficReq)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected successful query_traffic response, got: %+v", resp)
	}

	graphArgs, _ := json.Marshal(map[string]any{
		"name": "get_parameter_graph",
		"arguments": map[string]any{
			"target_id": target.ID,
		},
	})
	callGraphReq := []byte(`{"jsonrpc":"2.0","id":6,"method":"tools/call","params":` + string(graphArgs) + `}`)
	resp = srv.HandleMessage(ctx, callGraphReq)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected successful get_parameter_graph response, got: %+v", resp)
	}

	// 6. Test run_assessment
	scanRepo := database.NewScanRepo(db)
	assessmentEngine := assessment.NewAssessmentEngine(findingRepo, nil)
	srv.SetAssessmentEngine(assessmentEngine, scanRepo)

	_ = trafficRepo.Insert(&models.TrafficRecord{
		TargetID:           target.ID,
		Method:             "GET",
		URL:                "https://test.com/login?token=abc",
		Path:               "/login",
		Host:               "test.com",
		ResponseStatusCode: 200,
		Category:           "auth",
	})

	assessArgs, _ := json.Marshal(map[string]any{
		"name": "run_assessment",
		"arguments": map[string]any{
			"target_id": target.ID,
		},
	})
	callAssessReq := []byte(`{"jsonrpc":"2.0","id":7,"method":"tools/call","params":` + string(assessArgs) + `}`)
	resp = srv.HandleMessage(ctx, callAssessReq)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected successful run_assessment response, got: %+v", resp)
	}

	// 7. Test list_network_hosts and list_network_services
	hostRepo := database.NewHostRepo(db)
	serviceRepo := database.NewServiceRepo(db)
	srv.SetNetworkRepos(hostRepo, serviceRepo)

	_ = hostRepo.Upsert(&models.Host{
		TargetID: target.ID,
		IP:       "192.168.1.1",
		Status:   "up",
	})

	netArgs, _ := json.Marshal(map[string]any{
		"name": "list_network_hosts",
		"arguments": map[string]any{
			"target_id": target.ID,
		},
	})
	callNetReq := []byte(`{"jsonrpc":"2.0","id":8,"method":"tools/call","params":` + string(netArgs) + `}`)
	resp = srv.HandleMessage(ctx, callNetReq)
	if resp == nil || resp.Error != nil {
		t.Fatalf("expected successful list_network_hosts response, got: %+v", resp)
	}
}
