package tools

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// TheHarvesterJSON represents the JSON output format of theHarvester.
type TheHarvesterJSON struct {
	Hosts []string `json:"hosts"`
	IPs   []string `json:"ips"`
}

// TheHarvesterTool implements the Tool interface for theHarvester OSINT reconnaissance.
type TheHarvesterTool struct{}

// NewTheHarvesterTool creates a new TheHarvesterTool.
func NewTheHarvesterTool() *TheHarvesterTool {
	return &TheHarvesterTool{}
}

func (t *TheHarvesterTool) Name() string {
	return "theharvester"
}

func (t *TheHarvesterTool) BinaryName() string {
	if _, err := exec.LookPath("theHarvester"); err == nil {
		return "theHarvester"
	}
	if _, err := exec.LookPath("theharvester"); err == nil {
		return "theharvester"
	}
	return "uv"
}

func (t *TheHarvesterTool) Category() Category {
	return CategorySubdomainEnum
}

func (t *TheHarvesterTool) Description() string {
	return "Public OSINT gatherer querying certificate logs and search engines for subdomains."
}

func (t *TheHarvesterTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "theharvester_"+scanID)
}

func (t *TheHarvesterTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	clean := CleanDomain(target.Domain)
	// Highly reliable passive sources without requiring API keys
	sources := "crtsh,certspotter,hackertarget,otx,duckduckgo,rapiddns,urlscan"

	baseArgs := []string{
		"-d", clean,
		"-b", sources,
	}
	if outputFile != "" {
		baseArgs = append(baseArgs, "-f", outputFile)
	}

	if t.BinaryName() == "uv" {
		return append([]string{"run", "theHarvester"}, baseArgs...)
	}

	return baseArgs
}

func (t *TheHarvesterTool) IsInstalled() bool {
	_, err1 := exec.LookPath("theHarvester")
	_, err2 := exec.LookPath("theharvester")
	_, err3 := exec.LookPath("uv")
	return err1 == nil || err2 == nil || err3 == nil
}

func (t *TheHarvesterTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addHost := func(host string, ip string) {
		host = strings.TrimSpace(strings.ToLower(host))
		if idx := strings.Index(host, ":"); idx != -1 {
			if ip == "" {
				ip = host[idx+1:]
			}
			host = host[:idx]
		}
		if host == "" || seen[host] || !IsInScope(host, cleanRoot) {
			return
		}
		seen[host] = true

		assets = append(assets, models.Asset{
			ID:        uuid.New().String(),
			TargetID:  execCtx.Target.ID,
			ScanID:    execCtx.Scan.ID,
			Host:      host,
			IP:        ip,
			CreatedAt: now,
		})
	}

	// 1. Try reading the generated JSON file (theHarvester writes <outputFile>.json)
	if execCtx.OutputFile != "" {
		jsonPath := execCtx.OutputFile + ".json"
		if data, err := os.ReadFile(jsonPath); err == nil {
			var parsed TheHarvesterJSON
			if err := json.Unmarshal(data, &parsed); err == nil {
				for _, h := range parsed.Hosts {
					addHost(h, "")
				}
			}
			_ = os.Remove(jsonPath)
		}
		_ = os.Remove(execCtx.OutputFile + ".xml")
	}

	// 2. Parse stdout stream lines
	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "[*]") || strings.HasPrefix(line, "--") {
			continue
		}
		if strings.Contains(line, cleanRoot) && !strings.Contains(line, " ") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				addHost(parts[0], parts[1])
			} else {
				addHost(parts[0], "")
			}
		}
	}

	return Result{Assets: assets}, nil
}
