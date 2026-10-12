package tools

import (
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// GauTool implements the Tool interface for Corben Leo's gau (GetAllURLs).
type GauTool struct{}

// NewGauTool creates a new GauTool.
func NewGauTool() *GauTool {
	return &GauTool{}
}

func (t *GauTool) Name() string {
	return "gau"
}

func (t *GauTool) BinaryName() string {
	return "gau"
}

func (t *GauTool) Category() Category {
	return CategoryURLDiscovery
}

func (t *GauTool) Description() string {
	return "Fetches known URLs from AlienVault, Wayback Machine, and Common Crawl archives."
}

func (t *GauTool) OutputFile(scanID string) string {
	return ""
}

func (t *GauTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	clean := CleanDomain(target.Domain)
	return []string{"--subs", clean}
}

func (t *GauTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *GauTool) Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error) {
	var assets []models.Asset
	now := time.Now().UTC()

	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parsed, err := url.Parse(line)
		if err != nil || parsed.Hostname() == "" || !IsInScope(parsed.Hostname(), execCtx.Target.Domain) {
			continue
		}

		assets = append(assets, models.Asset{
			ID:        uuid.New().String(),
			TargetID:  execCtx.Target.ID,
			ScanID:    execCtx.Scan.ID,
			Host:      parsed.Hostname(),
			URL:       line,
			CreatedAt: now,
		})
	}

	return assets, nil
}
