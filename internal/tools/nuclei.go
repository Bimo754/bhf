package tools

import (
	"encoding/json"
	"net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// NucleiJSON represents a finding output from nuclei -jsonl.
type NucleiJSON struct {
	TemplateID string `json:"template-id"`
	Info       struct {
		Name     string   `json:"name"`
		Severity string   `json:"severity"`
		Tags     []string `json:"tags"`
	} `json:"info"`
	Host    string `json:"host"`
	Matched string `json:"matched-at"`
	Type    string `json:"type"`
	IP      string `json:"ip"`
}

// NucleiTool implements the Tool interface for ProjectDiscovery's Nuclei.
type NucleiTool struct{}

// NewNucleiTool creates a new NucleiTool.
func NewNucleiTool() *NucleiTool {
	return &NucleiTool{}
}

func (t *NucleiTool) Name() string {
	return "nuclei"
}

func (t *NucleiTool) BinaryName() string {
	return "nuclei"
}

func (t *NucleiTool) Category() Category {
	return CategoryVulnerability
}

func (t *NucleiTool) Description() string {
	return "Fast template-based vulnerability and exposure scanner."
}

func (t *NucleiTool) DefaultArgs(targetDomain string) []string {
	return []string{"-u", targetDomain, "-jsonl", "-silent", "-s", "info,low,medium,high,critical"}
}

func (t *NucleiTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *NucleiTool) Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error) {
	var assets []models.Asset
	now := time.Now().UTC()

	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var item NucleiJSON
		if err := json.Unmarshal([]byte(line), &item); err == nil && item.Host != "" {
			host := item.Host
			parsed, err := url.Parse(item.Host)
			if err == nil && parsed.Hostname() != "" {
				host = parsed.Hostname()
			}

			assets = append(assets, models.Asset{
				ID:           uuid.New().String(),
				TargetID:     execCtx.Target.ID,
				ScanID:       execCtx.Scan.ID,
				Host:         host,
				URL:          item.Matched,
				IP:           item.IP,
				Title:        item.Info.Name + " [" + strings.ToUpper(item.Info.Severity) + "]",
				Technologies: item.Info.Tags,
				CreatedAt:    now,
			})
		}
	}

	return assets, nil
}
