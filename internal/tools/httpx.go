package tools

import (
	"encoding/json"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// HttpxJSON represents the structured JSON output from httpx.
type HttpxJSON struct {
	URL        string   `json:"url"`
	Input      string   `json:"input"`
	Host       string   `json:"host"`
	Port       string   `json:"port"`
	Title      string   `json:"title"`
	StatusCode int      `json:"status_code"`
	Tech       []string `json:"tech"`
	IP         string   `json:"host"` // httpx sometimes puts IP here or in host
}

// HttpxTool implements the Tool interface for ProjectDiscovery's httpx.
type HttpxTool struct{}

// NewHttpxTool creates a new HttpxTool.
func NewHttpxTool() *HttpxTool {
	return &HttpxTool{}
}

func (t *HttpxTool) Name() string {
	return "httpx"
}

func (t *HttpxTool) BinaryName() string {
	return "httpx"
}

func (t *HttpxTool) Category() Category {
	return CategoryHTTPProbing
}

func (t *HttpxTool) Description() string {
	return "Fast multi-purpose HTTP probing toolkit for live service verification and technology detection."
}

func (t *HttpxTool) DefaultArgs(targetDomain string) []string {
	return []string{"-u", targetDomain, "-json", "-silent", "-tech-detect", "-title", "-status-code"}
}

func (t *HttpxTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *HttpxTool) Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error) {
	var assets []models.Asset
	now := time.Now().UTC()

	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var item HttpxJSON
		if err := json.Unmarshal([]byte(line), &item); err == nil && item.URL != "" {
			host := item.Input
			if host == "" {
				parsed, err := url.Parse(item.URL)
				if err == nil {
					host = parsed.Hostname()
				}
			}

			portNum := 0
			if item.Port != "" {
				portNum, _ = strconv.Atoi(item.Port)
			}

			assets = append(assets, models.Asset{
				ID:           uuid.New().String(),
				TargetID:     execCtx.Target.ID,
				ScanID:       execCtx.Scan.ID,
				Host:         host,
				URL:          item.URL,
				Port:         portNum,
				StatusCode:   item.StatusCode,
				Title:        item.Title,
				Technologies: item.Tech,
				CreatedAt:    now,
			})
		}
	}

	return assets, nil
}
