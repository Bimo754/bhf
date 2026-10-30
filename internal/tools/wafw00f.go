package tools

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// Wafw00fJSON represents an item in the JSON array output of wafw00f.
type Wafw00fJSON struct {
	URL          string `json:"url"`
	Detected     bool   `json:"detected"`
	Firewall     string `json:"firewall"`
	Manufacturer string `json:"manufacturer"`
}

// Wafw00fTool implements the Tool interface for wafw00f WAF fingerprinting.
type Wafw00fTool struct{}

// NewWafw00fTool creates a new Wafw00fTool.
func NewWafw00fTool() *Wafw00fTool {
	return &Wafw00fTool{}
}

func (t *Wafw00fTool) Name() string {
	return "wafw00f"
}

func (t *Wafw00fTool) BinaryName() string {
	if _, err := exec.LookPath("wafw00f"); err == nil {
		return "wafw00f"
	}
	return "uv"
}

func (t *Wafw00fTool) Category() Category {
	return CategoryHTTPProbing
}

func (t *Wafw00fTool) Description() string {
	return "Web Application Firewall (WAF) vendor detection and fingerprinting utility."
}

func (t *Wafw00fTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "wafw00f_"+scanID+".json")
}

func (t *Wafw00fTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	cleanRoot := CleanDomain(target.Domain)
	var targets []string

	if len(liveURLs) > 0 {
		for _, u := range liveURLs {
			u = strings.TrimSpace(u)
			if u != "" {
				targets = append(targets, u)
			}
		}
	} else if len(hosts) > 0 {
		for _, h := range hosts {
			h = strings.TrimSpace(h)
			if h != "" && IsInScope(h, cleanRoot) {
				targets = append(targets, "https://"+h)
			}
		}
	}

	if len(targets) == 0 {
		targets = append(targets, "https://"+cleanRoot)
	}

	var baseArgs []string
	listFile := outputFile + ".targets.txt"
	if err := os.WriteFile(listFile, []byte(strings.Join(targets, "\n")), 0644); err == nil {
		baseArgs = []string{
			"-i", listFile,
			"-o", outputFile,
			"-f", "json",
		}
	} else {
		baseArgs = []string{
			targets[0],
			"-o", outputFile,
			"-f", "json",
		}
	}

	if t.BinaryName() == "uv" {
		return append([]string{"run", "wafw00f"}, baseArgs...)
	}

	return baseArgs
}

func (t *Wafw00fTool) IsInstalled() bool {
	_, err1 := exec.LookPath("wafw00f")
	_, err2 := exec.LookPath("uv")
	return err1 == nil || err2 == nil
}

func (t *Wafw00fTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addWafAsset := func(targetURL string, firewall string) {
		targetURL = strings.TrimSpace(targetURL)
		if targetURL == "" || firewall == "" || strings.EqualFold(firewall, "None") || strings.EqualFold(firewall, "generic") {
			return
		}

		parsed, err := url.Parse(targetURL)
		if err != nil || parsed.Hostname() == "" {
			return
		}

		host := strings.ToLower(parsed.Hostname())
		if seen[host+firewall] || !IsInScope(host, cleanRoot) {
			return
		}
		seen[host+firewall] = true

		assets = append(assets, models.Asset{
			ID:           uuid.New().String(),
			TargetID:     execCtx.Target.ID,
			ScanID:       execCtx.Scan.ID,
			Host:         host,
			URL:          targetURL,
			Title:        fmt.Sprintf("WAF Protected: %s", firewall),
			Technologies: []string{"WAF:" + firewall},
			CreatedAt:    now,
		})
	}

	// 1. Read JSON file if available
	if execCtx.OutputFile != "" {
		if data, err := os.ReadFile(execCtx.OutputFile); err == nil && len(data) > 0 {
			var items []Wafw00fJSON
			if err := json.Unmarshal(data, &items); err == nil {
				for _, item := range items {
					if item.Detected && item.Firewall != "" {
						addWafAsset(item.URL, item.Firewall)
					}
				}
			}
			_ = os.Remove(execCtx.OutputFile)
		}
		_ = os.Remove(execCtx.OutputFile + ".targets.txt")
	}

	// 2. Parse stdout lines fallback
	if len(assets) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "is behind") {
				// Format: The site https://example.com is behind Cloudflare (Cloudflare Inc.) WAF.
				parts := strings.Fields(line)
				var detectedURL string
				for _, p := range parts {
					if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
						detectedURL = p
						break
					}
				}
				if detectedURL != "" {
					idx := strings.Index(line, "is behind")
					after := line[idx+len("is behind"):]
					after = strings.TrimSpace(after)
					wafName := strings.TrimSuffix(after, " WAF.")
					wafName = strings.TrimSuffix(wafName, " WAF")
					if pIdx := strings.Index(wafName, "("); pIdx != -1 {
						wafName = strings.TrimSpace(wafName[:pIdx])
					}
					addWafAsset(detectedURL, wafName)
				}
			}
		}
	}

	return Result{Assets: assets}, nil
}
