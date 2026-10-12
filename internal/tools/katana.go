package tools

import (
	"bufio"
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// KatanaJSON represents Katana crawler JSON output line.
type KatanaJSON struct {
	Endpoint string `json:"endpoint"`
	Request  struct {
		URL string `json:"url"`
	} `json:"request"`
	Response struct {
		StatusCode int `json:"status_code"`
	} `json:"response"`
}

// KatanaTool implements the Tool interface for ProjectDiscovery's Katana crawler.
type KatanaTool struct{}

// NewKatanaTool creates a new KatanaTool.
func NewKatanaTool() *KatanaTool {
	return &KatanaTool{}
}

func (t *KatanaTool) Name() string {
	return "katana"
}

func (t *KatanaTool) BinaryName() string {
	return "katana"
}

func (t *KatanaTool) Category() Category {
	return CategoryCrawling
}

func (t *KatanaTool) Description() string {
	return "Next-generation web application crawler to discover dynamic web endpoints and API routes."
}

func (t *KatanaTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "katana_"+scanID+".txt")
}

func (t *KatanaTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	targetURL := "https://" + CleanDomain(target.Domain)
	if len(liveURLs) > 0 {
		targetURL = liveURLs[0]
	}

	return []string{
		"-u", targetURL,
		"-depth", "2",
		"-timeout", "10",
		"-c", "10",
		"-rl", "100",
		"-iqp",
		"-o", outputFile,
	}
}

func (t *KatanaTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *KatanaTool) Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error) {
	var assets []models.Asset
	seen := make(map[string]bool)
	now := time.Now().UTC()

	parseURL := func(raw string, statusCode int) {
		raw = strings.TrimSpace(raw)
		if raw == "" || (!strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://")) {
			return
		}
		if seen[raw] {
			return
		}
		seen[raw] = true

		if parsed, err := url.Parse(raw); err == nil && parsed.Hostname() != "" {
			assets = append(assets, models.Asset{
				ID:         uuid.New().String(),
				TargetID:   execCtx.Target.ID,
				ScanID:     execCtx.Scan.ID,
				Host:       parsed.Hostname(),
				URL:        raw,
				StatusCode: statusCode,
				CreatedAt:  now,
			})
		}
	}

	// 1. Try reading the output file (handles both plain URLs and JSON lines)
	if execCtx.OutputFile != "" {
		if f, err := os.Open(execCtx.OutputFile); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				var item KatanaJSON
				if strings.HasPrefix(line, "{") && json.Unmarshal([]byte(line), &item) == nil {
					endpoint := item.Endpoint
					if endpoint == "" {
						endpoint = item.Request.URL
					}
					parseURL(endpoint, item.Response.StatusCode)
				} else {
					parseURL(line, 0)
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
		}
	}

	// 2. Fallback to parsing rawOutput lines from terminal stream
	if len(assets) == 0 {
		for _, line := range rawOutput {
			parseURL(line, 0)
		}
	}

	return assets, nil
}
