package api

import (
	"net/http"

	"github.com/diamond/bhf/internal/tools"
)

// ToolHandler provides endpoints for tool discovery and status.
type ToolHandler struct {
	registry *tools.Registry
}

// NewToolHandler creates a new ToolHandler.
func NewToolHandler(registry *tools.Registry) *ToolHandler {
	return &ToolHandler{registry: registry}
}

// ToolInfo represents details about a supported security tool.
type ToolInfo struct {
	Name        string         `json:"name"`
	BinaryName  string         `json:"binary_name"`
	Category    tools.Category `json:"category"`
	Description string         `json:"description"`
	IsInstalled bool           `json:"is_installed"`
}

// List returns metadata and installation status for all registered tools.
func (h *ToolHandler) List(w http.ResponseWriter, r *http.Request) {
	allTools := h.registry.List()
	var result []ToolInfo

	for _, t := range allTools {
		result = append(result, ToolInfo{
			Name:        t.Name(),
			BinaryName:  t.BinaryName(),
			Category:    t.Category(),
			Description: t.Description(),
			IsInstalled: t.IsInstalled(),
		})
	}

	JSON(w, http.StatusOK, result)
}
