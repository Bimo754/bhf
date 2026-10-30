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

// ArjunTool implements the Tool interface for Arjun parameter discovery.
type ArjunTool struct{}

// NewArjunTool creates a new ArjunTool.
func NewArjunTool() *ArjunTool {
	return &ArjunTool{}
}

func (t *ArjunTool) Name() string {
	return "arjun"
}

func (t *ArjunTool) BinaryName() string {
	if _, err := exec.LookPath("arjun"); err == nil {
		return "arjun"
	}
	return "uv"
}

func (t *ArjunTool) Category() Category {
	return CategoryFuzzing
}

func (t *ArjunTool) Description() string {
	return "HTTP parameter discovery suite detecting hidden GET, POST, and JSON parameters."
}

func (t *ArjunTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "arjun_"+scanID+".json")
}

func (t *ArjunTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
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
			"-t", "5",
			"-m", "GET",
			"-oJ", outputFile,
		}
	} else {
		baseArgs = []string{
			"-u", targets[0],
			"-t", "5",
			"-m", "GET",
			"-oJ", outputFile,
		}
	}

	if t.BinaryName() == "uv" {
		return append([]string{"run", "arjun"}, baseArgs...)
	}

	return baseArgs
}

func (t *ArjunTool) IsInstalled() bool {
	_, err1 := exec.LookPath("arjun")
	_, err2 := exec.LookPath("uv")
	return err1 == nil || err2 == nil
}

func (t *ArjunTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	var findings []models.Finding
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addParamFinding := func(rawURL string, params []string) {
		if rawURL == "" || len(params) == 0 {
			return
		}
		parsed, err := url.Parse(rawURL)
		if err != nil || parsed.Hostname() == "" {
			return
		}
		host := strings.ToLower(parsed.Hostname())
		if !IsInScope(host, cleanRoot) {
			return
		}

		key := rawURL + ":" + strings.Join(params, ",")
		if seen[key] {
			return
		}
		seen[key] = true

		paramListStr := strings.Join(params, ", ")
		sampleParam := params[0]
		curlCmd := fmt.Sprintf("curl -ivsL \"%s?%s=FUZZ\"", rawURL, sampleParam)

		findings = append(findings, models.Finding{
			ID:               uuid.New().String(),
			TargetID:         execCtx.Target.ID,
			ScanID:           execCtx.Scan.ID,
			ToolName:         t.Name(),
			Severity:         models.SeverityMedium,
			Name:             fmt.Sprintf("Discovered Hidden Parameters [%s]", paramListStr),
			MatchedAt:        rawURL,
			Description:      fmt.Sprintf("Hidden query parameter(s) [%s] detected on endpoint: %s", paramListStr, rawURL),
			ExtractedResults: params,
			CurlCommand:      curlCmd,
			CreatedAt:        now,
		})

		assets = append(assets, models.Asset{
			ID:           uuid.New().String(),
			TargetID:     execCtx.Target.ID,
			ScanID:       execCtx.Scan.ID,
			Host:         host,
			URL:          rawURL,
			Title:        fmt.Sprintf("Params: %s", paramListStr),
			Technologies: []string{"Param:" + sampleParam},
			CreatedAt:    now,
		})
	}

	// 1. Read JSON file if available
	if execCtx.OutputFile != "" {
		if data, err := os.ReadFile(execCtx.OutputFile); err == nil && len(data) > 0 {
			var parsed map[string][]string
			if err := json.Unmarshal(data, &parsed); err == nil {
				for endpoint, params := range parsed {
					addParamFinding(endpoint, params)
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
			if strings.Contains(line, "Parameters found for") {
				// Format: [+] Parameters found for https://target.com/api: id, debug
				idx := strings.Index(line, "Parameters found for")
				after := strings.TrimSpace(line[idx+len("Parameters found for"):])
				lastColon := strings.LastIndex(after, ":")
				if lastColon != -1 {
					endpointURL := strings.TrimSpace(after[:lastColon])
					rawParams := strings.Split(after[lastColon+1:], ",")
					var cleanParams []string
					for _, p := range rawParams {
						p = strings.TrimSpace(p)
						if p != "" {
							cleanParams = append(cleanParams, p)
						}
					}
					addParamFinding(endpointURL, cleanParams)
				}
			}
		}
	}

	return Result{
		Assets:   assets,
		Findings: findings,
	}, nil
}
