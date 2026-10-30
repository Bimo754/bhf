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

// CorsyJSON represents a finding item from corsy JSON output.
type CorsyJSON struct {
	URL         string `json:"url"`
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Curl        string `json:"curl"`
}

// CorsyTool implements the Tool interface for Corsy CORS misconfiguration scanner.
type CorsyTool struct{}

// NewCorsyTool creates a new CorsyTool.
func NewCorsyTool() *CorsyTool {
	return &CorsyTool{}
}

func (t *CorsyTool) Name() string {
	return "corsy"
}

func (t *CorsyTool) BinaryName() string {
	if _, err := exec.LookPath("corsy"); err == nil {
		return "corsy"
	}
	return "uv"
}

func (t *CorsyTool) Category() Category {
	return CategoryVulnerability
}

func (t *CorsyTool) Description() string {
	return "CORS misconfiguration scanner detecting arbitrary origin reflection with credentials."
}

func (t *CorsyTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "corsy_"+scanID+".json")
}

func (t *CorsyTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
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
	if err := os.WriteFile(listFile, []byte(strings.Join(targets, "\n")), 0644); err == nil && len(targets) > 1 {
		baseArgs = []string{
			"-u", targets[0],
			"-i", listFile,
			"-t", "10",
			"-q",
			"-o", outputFile,
		}
	} else {
		baseArgs = []string{
			"-u", targets[0],
			"-t", "10",
			"-q",
			"-o", outputFile,
		}
	}

	if t.BinaryName() == "uv" {
		return append([]string{"run", "corsy"}, baseArgs...)
	}

	return baseArgs
}

func (t *CorsyTool) IsInstalled() bool {
	_, err1 := exec.LookPath("corsy")
	_, err2 := exec.LookPath("uv")
	return err1 == nil || err2 == nil
}

func (t *CorsyTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	var findings []models.Finding
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addCorsFinding := func(targetURL string, vulnType string, sevStr string, desc string, curlCmd string) {
		targetURL = strings.TrimSpace(targetURL)
		if targetURL == "" {
			return
		}

		parsed, err := url.Parse(targetURL)
		if err != nil || parsed.Hostname() == "" {
			return
		}

		host := strings.ToLower(parsed.Hostname())
		if !IsInScope(host, cleanRoot) || seen[targetURL] {
			return
		}
		seen[targetURL] = true

		if vulnType == "" {
			vulnType = "Insecure CORS Policy"
		}
		if desc == "" {
			desc = "Arbitrary origin is accepted in Access-Control-Allow-Origin response header."
		}
		if curlCmd == "" {
			curlCmd = fmt.Sprintf("curl -ivsL -H \"Origin: https://evil.com\" \"%s\"", targetURL)
		}

		sev := models.SeverityMedium
		if strings.EqualFold(sevStr, "high") || strings.Contains(strings.ToLower(desc), "credential") {
			sev = models.SeverityHigh
		}

		findings = append(findings, models.Finding{
			ID:          uuid.New().String(),
			TargetID:    execCtx.Target.ID,
			ScanID:      execCtx.Scan.ID,
			ToolName:    t.Name(),
			Severity:    sev,
			Name:        fmt.Sprintf("CORS Misconfiguration - %s", vulnType),
			MatchedAt:   targetURL,
			Description: desc,
			CurlCommand: curlCmd,
			CreatedAt:   now,
		})

		assets = append(assets, models.Asset{
			ID:            uuid.New().String(),
			TargetID:      execCtx.Target.ID,
			ScanID:        execCtx.Scan.ID,
			Host:          host,
			URL:           targetURL,
			Title:         fmt.Sprintf("VULNERABLE: CORS Misconfiguration [%s]", vulnType),
			PriorityScore: 40,
			CreatedAt:     now,
		})
	}

	// 1. Read JSON file if available
	if execCtx.OutputFile != "" {
		if data, err := os.ReadFile(execCtx.OutputFile); err == nil && len(data) > 0 {
			var items []CorsyJSON
			if err := json.Unmarshal(data, &items); err == nil {
				for _, item := range items {
					addCorsFinding(item.URL, item.Type, item.Severity, item.Description, item.Curl)
				}
			}
			_ = os.Remove(execCtx.OutputFile)
		}
		_ = os.Remove(execCtx.OutputFile + ".targets.txt")
	}

	// 2. Parse stdout lines fallback
	if len(findings) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Misconfiguration") || strings.Contains(line, "ACAC: true") || strings.Contains(line, "Origin Reflection") {
				fields := strings.Fields(line)
				var detectedURL string
				for _, f := range fields {
					if strings.HasPrefix(f, "http://") || strings.HasPrefix(f, "https://") {
						detectedURL = strings.Trim(f, "[](),")
						break
					}
				}
				if detectedURL != "" {
					addCorsFinding(detectedURL, "Dynamic Origin Reflection", "High", line, "")
				}
			}
		}
	}

	return Result{
		Assets:   assets,
		Findings: findings,
	}, nil
}
