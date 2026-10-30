package tools

import (
	"bufio"
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

// JsluiceURLJSON represents a URL extraction item from jsluice.
type JsluiceURLJSON struct {
	URL      string `json:"url"`
	Type     string `json:"type"`
	Filename string `json:"filename"`
}

// JsluiceSecretJSON represents a secret extraction item from jsluice.
type JsluiceSecretJSON struct {
	Kind        string `json:"kind"`
	Match       string `json:"match"`
	Filename    string `json:"filename"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// JsluiceTool implements the Tool interface for Bishop Fox's jsluice.
type JsluiceTool struct{}

// NewJsluiceTool creates a new JsluiceTool.
func NewJsluiceTool() *JsluiceTool {
	return &JsluiceTool{}
}

func (t *JsluiceTool) Name() string {
	return "jsluice"
}

func (t *JsluiceTool) BinaryName() string {
	return "jsluice"
}

func (t *JsluiceTool) Category() Category {
	return CategoryURLDiscovery
}

func (t *JsluiceTool) Description() string {
	return "Tree-sitter AST JavaScript parser for extracting hidden API routes and leaked secrets."
}

func (t *JsluiceTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "jsluice_"+scanID+".jsonl")
}

func (t *JsluiceTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	cleanRoot := CleanDomain(target.Domain)
	var jsTargets []string

	// Prioritize JavaScript URLs
	for _, u := range liveURLs {
		u = strings.TrimSpace(u)
		if strings.Contains(u, ".js") {
			jsTargets = append(jsTargets, u)
		}
	}

	// Fallback to general live URLs if no .js paths found
	if len(jsTargets) == 0 && len(liveURLs) > 0 {
		jsTargets = append(jsTargets, liveURLs...)
	}

	if len(jsTargets) == 0 {
		jsTargets = append(jsTargets, "https://"+cleanRoot)
	}

	args := []string{
		"urls",
		"--resolve-paths", "https://" + cleanRoot,
	}

	// Limit targets to first 50 to avoid overly long command lines
	if len(jsTargets) > 50 {
		jsTargets = jsTargets[:50]
	}

	return append(args, jsTargets...)
}

func (t *JsluiceTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *JsluiceTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	var findings []models.Finding
	seenAssets := make(map[string]bool)
	seenFindings := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	parseLine := func(line string) {
		line = strings.TrimSpace(line)
		if line == "" || !strings.HasPrefix(line, "{") {
			return
		}

		// 1. Try URL extraction format
		var urlItem JsluiceURLJSON
		if err := json.Unmarshal([]byte(line), &urlItem); err == nil && urlItem.URL != "" {
			extractedURL := urlItem.URL
			if strings.HasPrefix(extractedURL, "/") {
				extractedURL = "https://" + cleanRoot + extractedURL
			}

			if parsed, err := url.Parse(extractedURL); err == nil && parsed.Hostname() != "" {
				host := strings.ToLower(parsed.Hostname())
				if IsInScope(host, cleanRoot) && !seenAssets[extractedURL] {
					seenAssets[extractedURL] = true
					assets = append(assets, models.Asset{
						ID:           uuid.New().String(),
						TargetID:     execCtx.Target.ID,
						ScanID:       execCtx.Scan.ID,
						Host:         host,
						URL:          extractedURL,
						Title:        "Discovered JS Route",
						Technologies: []string{"JavaScript Route"},
						CreatedAt:    now,
					})
				}
			}
			return
		}

		// 2. Try Secret extraction format
		var secretItem JsluiceSecretJSON
		if err := json.Unmarshal([]byte(line), &secretItem); err == nil && (secretItem.Kind != "" || secretItem.Match != "") {
			kind := secretItem.Kind
			if kind == "" {
				kind = "Generic Secret"
			}
			key := kind + ":" + secretItem.Match
			if seenFindings[key] {
				return
			}
			seenFindings[key] = true

			sev := models.SeverityHigh
			if strings.Contains(strings.ToLower(kind), "aws") || strings.Contains(strings.ToLower(kind), "private") {
				sev = models.SeverityCritical
			}

			curlCmd := fmt.Sprintf("curl -ivsL \"%s\"", secretItem.Filename)
			findings = append(findings, models.Finding{
				ID:               uuid.New().String(),
				TargetID:         execCtx.Target.ID,
				ScanID:           execCtx.Scan.ID,
				ToolName:         t.Name(),
				Severity:         sev,
				Name:             fmt.Sprintf("Leaked Secret [%s]", kind),
				MatchedAt:        secretItem.Filename,
				Description:      fmt.Sprintf("Hardcoded secret discovered via JavaScript AST analysis: %s", kind),
				ExtractedResults: []string{secretItem.Match},
				CurlCommand:      curlCmd,
				CreatedAt:        now,
			})
		}
	}

	// 1. Parse OutputFile if present
	if execCtx.OutputFile != "" {
		if f, err := os.Open(execCtx.OutputFile); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				parseLine(scanner.Text())
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
		}
		_ = os.Remove(execCtx.OutputFile + ".targets.txt")
	}

	// 2. Parse raw output lines
	for _, line := range rawOutput {
		parseLine(line)
	}

	return Result{
		Assets:   assets,
		Findings: findings,
	}, nil
}
