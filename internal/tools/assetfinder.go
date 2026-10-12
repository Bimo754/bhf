package tools

import (
	"os/exec"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// AssetfinderTool implements the Tool interface for TomNomNom's assetfinder.
type AssetfinderTool struct{}

// NewAssetfinderTool creates a new AssetfinderTool.
func NewAssetfinderTool() *AssetfinderTool {
	return &AssetfinderTool{}
}

func (t *AssetfinderTool) Name() string {
	return "assetfinder"
}

func (t *AssetfinderTool) BinaryName() string {
	return "assetfinder"
}

func (t *AssetfinderTool) Category() Category {
	return CategorySubdomainEnum
}

func (t *AssetfinderTool) Description() string {
	return "Finds related domains and subdomains using public sources and threat intelligence feeds."
}

func (t *AssetfinderTool) OutputFile(scanID string) string {
	return ""
}

func (t *AssetfinderTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	clean := CleanDomain(target.Domain)
	return []string{"--subs-only", clean}
}

func (t *AssetfinderTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *AssetfinderTool) Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error) {
	var assets []models.Asset
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "[") {
			continue
		}

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
