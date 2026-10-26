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

	"github.com/diamond/bhf/internal/database"
	"github.com/diamond/bhf/internal/tools"
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
	targetRepo  *database.TargetRepo
	assetRepo   *database.AssetRepo
	findingRepo *database.FindingRepo
	registry    *tools.Registry
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
		assets, _, err := s.assetRepo.ListByTarget(targetID, "", "", 1000, 0)
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

	default:
		return "", fmt.Errorf("unrecognized tool: %s", name)
	}
}
