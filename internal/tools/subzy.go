package tools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// SubzyJSONItem represents a single result from subzy JSON output.
type SubzyJSONItem struct {
	Subdomain  string `json:"subdomain"`
	Vulnerable bool   `json:"vulnerable"`
	Service    string `json:"service"`
	Status     string `json:"status"`
	Error      string `json:"error"`
}

// SubzyTool implements the Tool interface for subzy subdomain takeover detection.
type SubzyTool struct{}

// NewSubzyTool creates a new SubzyTool.
func NewSubzyTool() *SubzyTool {
	return &SubzyTool{}
}

func (t *SubzyTool) Name() string {
	return "subzy"
}

func (t *SubzyTool) BinaryName() string {
	return "subzy"
}

func (t *SubzyTool) Category() Category {
	return CategoryVulnerability
}

func (t *SubzyTool) Description() string {
	return "Subdomain takeover vulnerability checker targeting CNAME records."
}

func (t *SubzyTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "subzy_"+scanID+".json")
}

func (t *SubzyTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	var inputTargets []string
	cleanRoot := CleanDomain(target.Domain)

	for _, h := range hosts {
		h = strings.TrimSpace(h)
		if h != "" && IsInScope(h, cleanRoot) {
			inputTargets = append(inputTargets, h)
		}
	}

	if len(inputTargets) == 0 {
		inputTargets = append(inputTargets, cleanRoot)
	}

	listFile := outputFile + ".targets.txt"
	if err := os.WriteFile(listFile, []byte(strings.Join(inputTargets, "\n")), 0644); err == nil {
		return []string{
			"run",
			"--targets", listFile,
			"--concurrency", "25",
			"--hide_fails",
			"--output", outputFile,
		}
	}

	return []string{
		"run",
		"--target", cleanRoot,
		"--hide_fails",
	}
}

func (t *SubzyTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *SubzyTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var findings []models.Finding
	var assets []models.Asset
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addTakeover := func(host string, service string) {
		host = strings.TrimSpace(strings.ToLower(host))
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
		if host == "" || seen[host] || !IsInScope(host, cleanRoot) {
			return
		}
		seen[host] = true

		if service == "" {
			service = "Unclaimed CNAME"
		}

		findingName := fmt.Sprintf("Subdomain Takeover - %s", service)
		curlCmd := fmt.Sprintf("curl -ivsL \"https://%s\"", host)

		findings = append(findings, models.Finding{
			ID:          uuid.New().String(),
			TargetID:    execCtx.Target.ID,
			ScanID:      execCtx.Scan.ID,
			ToolName:    t.Name(),
			Severity:    models.SeverityHigh,
			Name:        findingName,
			MatchedAt:   host,
			Description: fmt.Sprintf("Dangling CNAME record pointing to unclaimed external service: %s", service),
			CurlCommand: curlCmd,
			CreatedAt:   now,
		})

		assets = append(assets, models.Asset{
			ID:            uuid.New().String(),
			TargetID:      execCtx.Target.ID,
			ScanID:        execCtx.Scan.ID,
			Host:          host,
			Title:         fmt.Sprintf("VULNERABLE: Subdomain Takeover [%s]", service),
			PriorityScore: 50,
			CreatedAt:     now,
		})
	}

	// 1. Parse JSON output file if available
	if execCtx.OutputFile != "" {
		if data, err := os.ReadFile(execCtx.OutputFile); err == nil && len(data) > 0 {
			// Try as JSON array
			var items []SubzyJSONItem
			if err := json.Unmarshal(data, &items); err == nil {
				for _, item := range items {
					if item.Vulnerable || strings.EqualFold(item.Status, "vulnerable") {
						addTakeover(item.Subdomain, item.Service)
					}
				}
			} else {
				// Try line-by-line JSON
				scanner := bufio.NewScanner(strings.NewReader(string(data)))
				for scanner.Scan() {
					line := strings.TrimSpace(scanner.Text())
					var item SubzyJSONItem
					if json.Unmarshal([]byte(line), &item) == nil {
						if item.Vulnerable || strings.EqualFold(item.Status, "vulnerable") {
							addTakeover(item.Subdomain, item.Service)
						}
					}
				}
			}
			_ = os.Remove(execCtx.OutputFile)
		}
		_ = os.Remove(execCtx.OutputFile + ".targets.txt")
	}

	// 2. Parse stdout lines fallback
	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		upper := strings.ToUpper(line)
		if strings.Contains(upper, "NOT VULNERABLE") || strings.Contains(upper, "NOT_VULNERABLE") {
			continue
		}
		if strings.Contains(upper, "VULNERABLE") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if strings.Contains(p, cleanRoot) {
					service := ""
					if i+1 < len(parts) {
						service = strings.Trim(parts[i+1], "[]()-")
					}
					addTakeover(p, service)
					break
				}
			}
		}
	}

	return Result{
		Assets:   assets,
		Findings: findings,
	}, nil
}
