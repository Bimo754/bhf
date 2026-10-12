package tools

import (
	"encoding/json"
	"os/exec"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// SubfinderJSON represents a single JSON output item from subfinder -json.
type SubfinderJSON struct {
	Host   string `json:"host"`
	Source string `json:"source"`
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
	return "Fast passive subdomain enumeration tool using multiple passive online sources."
}

func (t *SubfinderTool) DefaultArgs(targetDomain string) []string {
	return []string{"-d", targetDomain, "-json", "-silent"}
}

func (t *SubfinderTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *SubfinderTool) Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error) {
	var assets []models.Asset
	now := time.Now().UTC()

	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var item SubfinderJSON
		if err := json.Unmarshal([]byte(line), &item); err == nil && item.Host != "" {
			assets = append(assets, models.Asset{
				ID:        uuid.New().String(),
				TargetID:  execCtx.Target.ID,
				ScanID:    execCtx.Scan.ID,
				Host:      item.Host,
				CreatedAt: now,
			})
		} else if !strings.HasPrefix(line, "{") {
			// Plain text fallback
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
