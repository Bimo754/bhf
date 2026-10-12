package tools

import (
	"encoding/json"
	"os/exec"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// SubfinderJSON represents a single JSON output item from subfinder.
type SubfinderJSON struct {
	Host string `json:"host"`
}

// SubfinderTool implements the Tool interface for ProjectDiscovery's Subfinder.
type SubfinderTool struct{}

// NewSubfinderTool creates a new SubfinderTool.
func NewSubfinderTool() *SubfinderTool {
	return &SubfinderTool{}
}

func (t *SubfinderTool) Name() string {
	return "subfinder"
}

func (t *SubfinderTool) BinaryName() string {
	return "subfinder"
}

func (t *SubfinderTool) Category() Category {
	return CategorySubdomainEnum
}

func (t *SubfinderTool) Description() string {
	return "Fast passive subdomain enumeration tool querying multiple online public sources."
}

func (t *SubfinderTool) OutputFile(scanID string) string {
	return ""
}

func (t *SubfinderTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	clean := CleanDomain(target.Domain)
	// Output live findings line-by-line to stdout without silencing
	return []string{"-d", clean}
}

func (t *SubfinderTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *SubfinderTool) Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error) {
	var assets []models.Asset
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "[") || strings.HasPrefix(line, "_") {
			continue
		}

		// Check if it's JSON format
		var item SubfinderJSON
		if strings.HasPrefix(line, "{") && json.Unmarshal([]byte(line), &item) == nil && item.Host != "" {
			assets = append(assets, models.Asset{
				ID:        uuid.New().String(),
				TargetID:  execCtx.Target.ID,
				ScanID:    execCtx.Scan.ID,
				Host:      item.Host,
				CreatedAt: now,
			})
			continue
		}

		// Plain text line
		if strings.Contains(line, cleanRoot) && !strings.Contains(line, " ") {
			assets = append(assets, models.Asset{
				ID:        uuid.New().String(),
				TargetID:  execCtx.Target.ID,
				ScanID:    execCtx.Scan.ID,
				Host:      line,
				CreatedAt: now,
			})
		}
	}

	return assets, nil
}
