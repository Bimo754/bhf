package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/diamond/bhf/internal/appintel"
	"github.com/diamond/bhf/internal/assessment"
	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/models"
	"github.com/diamond/bhf/internal/tools"
	"github.com/google/uuid"
)

// JSONRPCRequest represents an incoming JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents an outgoing JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      any           `json:"id,omitempty"`
	Result  any           `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error payload.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ToolDefinition defines an MCP tool's metadata and schema.
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// Server exposes BHF assets and security findings over the Model Context Protocol.
type Server struct {
	targetRepo       *database.TargetRepo
	assetRepo        *database.AssetRepo
	findingRepo      *database.FindingRepo
	trafficRepo      *database.TrafficRepo
	hostRepo         *database.HostRepo
	serviceRepo      *database.ServiceRepo
	scanRepo         *database.ScanRepo
	assessmentEngine *assessment.AssessmentEngine
	registry         *tools.Registry
}

// NewServer initializes a new MCP Server instance.
func NewServer(
	targetRepo *database.TargetRepo,
	assetRepo *database.AssetRepo,
	findingRepo *database.FindingRepo,
	registry *tools.Registry,
) *Server {
	return &Server{
		targetRepo:  targetRepo,
		assetRepo:   assetRepo,
		findingRepo: findingRepo,
		registry:    registry,
	}
}

// SetTrafficRepo binds the traffic repository for application intelligence queries.
func (s *Server) SetTrafficRepo(repo *database.TrafficRepo) {
	s.trafficRepo = repo
}

// SetAssessmentEngine binds the assessment engine and scan repository for deterministic audits.
func (s *Server) SetAssessmentEngine(engine *assessment.AssessmentEngine, scanRepo *database.ScanRepo) {
	s.assessmentEngine = engine
	s.scanRepo = scanRepo
}

// SetNetworkRepos binds the host and service repositories for network asset queries.
func (s *Server) SetNetworkRepos(hostRepo *database.HostRepo, serviceRepo *database.ServiceRepo) {
	s.hostRepo = hostRepo
	s.serviceRepo = serviceRepo
}

// ServeHTTP handles MCP requests over HTTP (e.g. POST /api/mcp).
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	resp := s.HandleMessage(r.Context(), body)
	if resp == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// RunStdio starts a JSON-RPC 2.0 stdio loop for desktop or command-line MCP clients.
func (s *Server) RunStdio(ctx context.Context) error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		resp := s.HandleMessage(ctx, line)
		if resp != nil {
			out, err := json.Marshal(resp)
			if err == nil {
				fmt.Println(string(out))
			}
		}
	}
	return scanner.Err()
}

// HandleMessage parses a raw JSON-RPC request and executes the appropriate handler.
func (s *Server) HandleMessage(ctx context.Context, raw []byte) *JSONRPCResponse {
	var req JSONRPCRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			Error: &JSONRPCError{
				Code:    -32700,
				Message: "Parse error",
			},
		}
	}

	switch req.Method {
	case "initialize":
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "bhf-mcp-server",
					"version": "1.0.0",
				},
			},
		}

	case "notifications/initialized":
		return nil

	case "tools/list":
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"tools": s.getToolDefinitions(),
			},
		}

	case "tools/call":
		var callParams struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &callParams); err != nil {
			return &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &JSONRPCError{
					Code:    -32602,
					Message: "Invalid params",
				},
			}
		}

		resultText, err := s.callTool(callParams.Name, callParams.Arguments)
		if err != nil {
			return &JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: map[string]any{
					"content": []map[string]any{
						{
							"type": "text",
							"text": fmt.Sprintf("Error: %v", err),
						},
					},
					"isError": true,
				},
			}
		}

		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"content": []map[string]any{
					{
						"type": "text",
						"text": resultText,
					},
				},
			},
		}

	default:
		return &JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &JSONRPCError{
				Code:    -32601,
				Message: fmt.Sprintf("Method '%s' not found", req.Method),
			},
		}
	}
}

func (s *Server) getToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "list_targets",
			Description: "List all configured bug bounty targets and scopes in BHF",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "list_assets",
			Description: "List discovered subdomains, IPs, open ports, and live URLs for a target scope",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_id": map[string]any{
						"type":        "string",
						"description": "The unique UUID of the target",
					},
				},
				"required": []string{"target_id"},
			},
		},
		{
			Name:        "list_findings",
			Description: "List discovered vulnerabilities and security findings for a target scope",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_id": map[string]any{
						"type":        "string",
						"description": "The unique UUID of the target",
					},
				},
				"required": []string{"target_id"},
			},
		},
		{
			Name:        "get_target_stats",
			Description: "Get asset, finding, and priority statistics for a target scope",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_id": map[string]any{
						"type":        "string",
						"description": "The unique UUID of the target",
					},
				},
				"required": []string{"target_id"},
			},
		},
		{
			Name:        "list_tools",
			Description: "List all available reconnaissance tools installed in BHF",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "query_traffic",
			Description: "Query intercepted HTTP transactions and requests by category, method, or search",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_id": map[string]any{
						"type":        "string",
						"description": "The unique UUID of the target",
					},
					"category": map[string]any{
						"type":        "string",
						"description": "Optional category filter (auth, api, upload, search, admin)",
					},
					"method": map[string]any{
						"type":        "string",
						"description": "Optional HTTP method filter (GET, POST, etc)",
					},
					"search": map[string]any{
						"type":        "string",
						"description": "Optional URL/path search query",
					},
				},
				"required": []string{"target_id"},
			},
		},
		{
			Name:        "get_parameter_graph",
			Description: "Retrieve aggregated user-controlled parameter graph and IDOR candidates for a target",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_id": map[string]any{
						"type":        "string",
						"description": "The unique UUID of the target",
					},
				},
				"required": []string{"target_id"},
			},
		},
		{
			Name:        "run_assessment",
			Description: "Execute deterministic security assessment rules (auth checks, IDOR, Canary reflection, upload) against captured HTTP traffic",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_id": map[string]any{
						"type":        "string",
						"description": "The unique UUID of the target",
					},
				},
				"required": []string{"target_id"},
			},
		},
		{
			Name:        "list_network_hosts",
			Description: "List discovered network hosts and IP addresses for a target scope",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_id": map[string]any{
						"type":        "string",
						"description": "The unique UUID of the target",
					},
				},
				"required": []string{"target_id"},
			},
		},
		{
			Name:        "list_network_services",
			Description: "List discovered network ports, services, and protocol banners for a target scope",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_id": map[string]any{
						"type":        "string",
						"description": "The unique UUID of the target",
					},
				},
				"required": []string{"target_id"},
			},
		},
		{
			Name:        "auto_explore_target",
			Description: "Autonomously explore, click around, submit forms, and record traffic for a target web application",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"target_id": map[string]any{
						"type":        "string",
						"description": "The unique UUID of the target",
					},
					"start_url": map[string]any{
						"type":        "string",
						"description": "Optional starting URL (defaults to https://<domain>)",
					},
					"max_pages": map[string]any{
						"type":        "integer",
						"description": "Maximum pages to visit (default 20)",
					},
				},
				"required": []string{"target_id"},
			},
		},
	}
}

func (s *Server) callTool(name string, args map[string]any) (string, error) {
	switch name {
	case "list_targets":
		targets, err := s.targetRepo.List()
		if err != nil {
			return "", err
		}
		out, err := json.MarshalIndent(targets, "", "  ")
		return string(out), err

	case "list_assets":
		targetID, _ := args["target_id"].(string)
		if targetID == "" {
			return "", fmt.Errorf("target_id is required")
		}
		assets, _, err := s.assetRepo.ListByTarget(targetID, "", "", "", "", 1000, 0)
		if err != nil {
			return "", err
		}
		out, err := json.MarshalIndent(assets, "", "  ")
		return string(out), err

	case "list_findings":
		targetID, _ := args["target_id"].(string)
		if targetID == "" {
			return "", fmt.Errorf("target_id is required")
		}
		findings, _, err := s.findingRepo.ListByTarget(targetID, "", "", 1000, 0)
		if err != nil {
			return "", err
		}
		out, err := json.MarshalIndent(findings, "", "  ")
		return string(out), err

	case "get_target_stats":
		targetID, _ := args["target_id"].(string)
		if targetID == "" {
			return "", fmt.Errorf("target_id is required")
		}
		assetStats, err := s.assetRepo.GetStats(targetID)
		if err != nil {
			log.Printf("[MCP] Asset stats error: %v", err)
		}
		findingStats, err := s.findingRepo.GetStats(targetID)
		if err != nil {
			log.Printf("[MCP] Finding stats error: %v", err)
		}
		combined := map[string]any{
			"assets":   assetStats,
			"findings": findingStats,
		}
		out, err := json.MarshalIndent(combined, "", "  ")
		return string(out), err

	case "list_tools":
		toolList := s.registry.List()
		out, err := json.MarshalIndent(toolList, "", "  ")
		return string(out), err

	case "query_traffic":
		if s.trafficRepo == nil {
			return "", fmt.Errorf("traffic repository is not configured")
		}
		targetID, _ := args["target_id"].(string)
		if targetID == "" {
			return "", fmt.Errorf("target_id is required")
		}
		category, _ := args["category"].(string)
		method, _ := args["method"].(string)
		search, _ := args["search"].(string)

		records, _, err := s.trafficRepo.ListByTarget(targetID, category, method, search, 50, 0)
		if err != nil {
			return "", err
		}
		out, err := json.MarshalIndent(records, "", "  ")
		return string(out), err

	case "get_parameter_graph":
		if s.trafficRepo == nil {
			return "", fmt.Errorf("traffic repository is not configured")
		}
		targetID, _ := args["target_id"].(string)
		if targetID == "" {
			return "", fmt.Errorf("target_id is required")
		}

		records, _, err := s.trafficRepo.ListByTarget(targetID, "", "", "", 200, 0)
		if err != nil {
			return "", err
		}
		graph := appintel.BuildParameterGraph(records)
		out, err := json.MarshalIndent(graph, "", "  ")
		return string(out), err

	case "run_assessment":
		if s.assessmentEngine == nil || s.trafficRepo == nil {
			return "", fmt.Errorf("assessment engine or traffic repository is not configured")
		}
		targetID, _ := args["target_id"].(string)
		if targetID == "" {
			return "", fmt.Errorf("target_id is required")
		}
		target, err := s.targetRepo.GetByID(targetID)
		if err != nil || target == nil {
			return "", fmt.Errorf("target not found")
		}
		records, _, err := s.trafficRepo.ListByTarget(targetID, "", "", "", 500, 0)
		if err != nil {
			return "", err
		}
		scanID := "mcp-scan"
		if s.scanRepo != nil {
			now := time.Now()
			sc := &models.Scan{
				ID:        uuid.New().String(),
				TargetID:  targetID,
				ToolName:  "mcp_assessment",
				Status:    models.StatusRunning,
				StartedAt: now,
				Arguments: []string{},
			}
			if err := s.scanRepo.Create(sc); err == nil {
				scanID = sc.ID
				defer func() {
					completedAt := time.Now()
					_ = s.scanRepo.UpdateStatus(sc.ID, models.StatusCompleted, &completedAt, "")
				}()
			}
		}
		findings, err := s.assessmentEngine.Run(context.Background(), target, records, scanID)
		if err != nil {
			return "", err
		}
		summary := map[string]any{
			"target_id":        targetID,
			"traffic_analyzed": len(records),
			"findings_count":   len(findings),
			"findings":         findings,
		}
		out, err := json.MarshalIndent(summary, "", "  ")
		return string(out), err

	case "list_network_hosts":
		if s.hostRepo == nil {
			return "", fmt.Errorf("network host repository is not configured")
		}
		targetID, _ := args["target_id"].(string)
		if targetID == "" {
			return "", fmt.Errorf("target_id is required")
		}
		hosts, _, err := s.hostRepo.ListByTarget(targetID, 100, 0)
		if err != nil {
			return "", err
		}
		out, err := json.MarshalIndent(hosts, "", "  ")
		return string(out), err

	case "list_network_services":
		if s.serviceRepo == nil {
			return "", fmt.Errorf("network service repository is not configured")
		}
		targetID, _ := args["target_id"].(string)
		if targetID == "" {
			return "", fmt.Errorf("target_id is required")
		}
		services, _, err := s.serviceRepo.ListByTarget(targetID, 100, 0)
		if err != nil {
			return "", err
		}
		out, err := json.MarshalIndent(services, "", "  ")
		return string(out), err

	case "auto_explore_target":
		if s.trafficRepo == nil {
			return "", fmt.Errorf("traffic repository is not configured")
		}
		targetID, _ := args["target_id"].(string)
		if targetID == "" {
			return "", fmt.Errorf("target_id is required")
		}
		target, err := s.targetRepo.GetByID(targetID)
		if err != nil || target == nil {
			return "", fmt.Errorf("target not found")
		}
		startURL, _ := args["start_url"].(string)
		maxPages := 20
		if mp, ok := args["max_pages"].(float64); ok && mp > 0 {
			maxPages = int(mp)
		}
		cfg := appintel.AutoExplorerConfig{
			Target:       target,
			TrafficRepo:  s.trafficRepo,
			ScanID:       "mcp-auto-explore",
			StartURL:     startURL,
			MaxPages:     maxPages,
			MaxDepth:     3,
			PreferChrome: true,
		}
		explorer := appintel.NewAutoExplorer(cfg, nil)
		progress, err := explorer.Run(context.Background())
		if err != nil && progress == nil {
			return "", err
		}
		out, err := json.MarshalIndent(progress, "", "  ")
		return string(out), err

	default:
		return "", fmt.Errorf("unrecognized tool: %s", name)
	}
}
