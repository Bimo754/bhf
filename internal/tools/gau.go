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

func (t *GauTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	var endpoints []models.Endpoint
	seenURLs := make(map[string]bool)
	seenHosts := make(map[string]bool)
	now := time.Now().UTC()

	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if line == "" || seenURLs[line] {
			continue
		}
		seenURLs[line] = true

		parsed, err := url.Parse(line)
		if err != nil || parsed.Hostname() == "" || !IsInScope(parsed.Hostname(), execCtx.Target.Domain) {
			continue
		}

		path := parsed.Path
		if path == "" {
			path = "/"
		}
		if parsed.RawQuery != "" {
			path += "?" + parsed.RawQuery
		}

		endpoints = append(endpoints, models.Endpoint{
			ID:         uuid.New().String(),
			TargetID:   execCtx.Target.ID,
			ScanID:     execCtx.Scan.ID,
			Host:       parsed.Hostname(),
			URL:        line,
			Path:       path,
			Method:     "GET",
			SourceTool: t.Name(),
			CreatedAt:  now,
		})

		if !seenHosts[parsed.Hostname()] {
			seenHosts[parsed.Hostname()] = true
			assets = append(assets, models.Asset{
				ID:        uuid.New().String(),
				TargetID:  execCtx.Target.ID,
				ScanID:    execCtx.Scan.ID,
				Host:      parsed.Hostname(),
				URL:       line,
				CreatedAt: now,
			})
		}
	}

	return Result{
		Assets:    assets,
		Endpoints: endpoints,
	}, nil
}
