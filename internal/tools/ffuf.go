package tools

import (
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// FfufJSON represents ffuf output structure.
type FfufJSON struct {
	Results []struct {
		Input    map[string]string `json:"input"`
		Position int               `json:"position"`
		Status   int               `json:"status"`
		Length   int               `json:"length"`
		URL      string            `json:"url"`
	} `json:"results"`
}

// FfufTool implements the Tool interface for ffuf web fuzzer.
type FfufTool struct{}

// NewFfufTool creates a new FfufTool.
func NewFfufTool() *FfufTool {
	return &FfufTool{}
}

func (t *FfufTool) Name() string {
	return "ffuf"
}

func (t *FfufTool) BinaryName() string {
	return "ffuf"
}

func (t *FfufTool) Category() Category {
	return CategoryFuzzing
}

func (t *FfufTool) Description() string {
	return "Fast web fuzzer for directory brute-forcing, virtual host discovery, and parameter analysis."
}

func (t *FfufTool) DefaultArgs(targetDomain string) []string {
	return []string{"-u", "https://" + targetDomain + "/FUZZ", "-mc", "200,204,301,302,307,401,403", "-s"}
}

func (t *FfufTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *FfufTool) Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error) {
	var assets []models.Asset
	now := time.Now().UTC()

	// Parse line by line or entire JSON
	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try ffuf result line
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			parsed, err := url.Parse(line)
			if err == nil {
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
	}

	return assets, nil
}
