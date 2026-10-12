package tools

import (
	"bufio"
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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

func (t *HttpxTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "httpx_"+scanID+".json")
}

func (t *HttpxTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	// If subdomains were already discovered in previous steps, probe all of them!
	if len(hosts) > 0 {
		listFile := filepath.Join(os.TempDir(), "httpx_targets_"+outputFile+".txt")
		_ = os.WriteFile(listFile, []byte(strings.Join(hosts, "\n")), 0644)
		return []string{"-l", listFile, "-status-code", "-title", "-tech-detect", "-o", outputFile, "-json"}
	}

	clean := CleanDomain(target.Domain)
	return []string{"-u", clean, "-status-code", "-title", "-tech-detect", "-o", outputFile, "-json"}
}

func (t *HttpxTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *HttpxTool) Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error) {
	var assets []models.Asset
	now := time.Now().UTC()

	// 1. Try reading the structured JSON file if it exists
	if execCtx.OutputFile != "" {
		if f, err := os.Open(execCtx.OutputFile); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				var item HttpxJSON
				if err := json.Unmarshal([]byte(line), &item); err == nil && item.URL != "" {
					host := item.Input
					if host == "" {
						if parsed, err := url.Parse(item.URL); err == nil {
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
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
		}
	}

	// 2. Fallback to parsing rawOutput lines if file was empty
	if len(assets) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
				parts := strings.Fields(line)
				rawURL := parts[0]
				if parsed, err := url.Parse(rawURL); err == nil {
					assets = append(assets, models.Asset{
						ID:        uuid.New().String(),
						TargetID:  execCtx.Target.ID,
						ScanID:    execCtx.Scan.ID,
						Host:      parsed.Hostname(),
						URL:       rawURL,
						CreatedAt: now,
					})
				}
			}
		}
	}

	return assets, nil
}
