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
	return "Next-generation crawling and spidering framework to discover dynamic web endpoints."
}

func (t *KatanaTool) DefaultArgs(targetDomain string) []string {
	return []string{"-u", "https://" + targetDomain, "-jsonl", "-silent", "-depth", "2"}
}

func (t *KatanaTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *KatanaTool) Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error) {
	var assets []models.Asset
	now := time.Now().UTC()

	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var item KatanaJSON
		if err := json.Unmarshal([]byte(line), &item); err == nil && (item.Endpoint != "" || item.Request.URL != "") {
			endpoint := item.Endpoint
			if endpoint == "" {
				endpoint = item.Request.URL
			}

			parsed, err := url.Parse(endpoint)
			if err != nil || parsed.Hostname() == "" {
				continue
			}

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

	return assets, nil
}
