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
	return filepath.Join(os.TempDir(), "katana_"+scanID+".jsonl")
}

func (t *KatanaTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	targetURL := "https://" + CleanDomain(target.Domain)
	if len(liveURLs) > 0 {
		targetURL = liveURLs[0]
	}

	return []string{
		"-u", targetURL,
		"-depth", "2",
		"-ct", "3m",
		"-timeout", "7",
		"-c", "10",
		"-rl", "50",
		"-iqp",
		"-o", outputFile,
		"-jsonl",
	}
}

func (t *KatanaTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *KatanaTool) Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error) {
	var assets []models.Asset
	now := time.Now().UTC()

	// 1. Try reading the structured JSON file
	if execCtx.OutputFile != "" {
		if f, err := os.Open(execCtx.OutputFile); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				var item KatanaJSON
				if err := json.Unmarshal([]byte(line), &item); err == nil {
					endpoint := item.Endpoint
					if endpoint == "" {
						endpoint = item.Request.URL
					}

					if parsed, err := url.Parse(endpoint); err == nil && parsed.Hostname() != "" {
						assets = append(assets, models.Asset{
							ID:         uuid.New().String(),
							TargetID:   execCtx.Target.ID,
							ScanID:     execCtx.Scan.ID,
							Host:       parsed.Hostname(),
							URL:        endpoint,
							StatusCode: item.Response.StatusCode,
							CreatedAt:  now,
						})
					}
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
		}
	}

	// 2. Fallback to parsing rawOutput lines
	if len(assets) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
				if parsed, err := url.Parse(line); err == nil && parsed.Hostname() != "" {
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
	}

	return assets, nil
}
