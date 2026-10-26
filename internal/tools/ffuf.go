package tools

import (
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

// FfufJSON represents ffuf JSON output structure.
type FfufJSON struct {
	Results []struct {
		URL    string `json:"url"`
		Status int    `json:"status"`
	} `json:"results"`
}

// FfufTool implements the Tool interface for ffuf web fuzzer.
type FfufTool struct{}

// NewFfufTool creates a new FfufTool.
func NewFfufTool() *FfufTool {
	return &FfufTool{}
}

func (t *FfufTool) Name() string {
	return "ffuf"
}

func (t *FfufTool) BinaryName() string {
	return "ffuf"
}

func (t *FfufTool) Category() Category {
	return CategoryFuzzing
}

func (t *FfufTool) Description() string {
	return "Fast web fuzzer for directory brute-forcing, virtual host discovery, and parameter analysis."
}

func (t *FfufTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "ffuf_"+scanID+".json")
}

func (t *FfufTool) ensureWordlist() string {
	// Common wordlist locations
	candidates := []string{
		"/app/wordlists/common.txt",
		"/usr/share/wordlists/dirb/common.txt",
		"/usr/share/wordlists/dirbuster/directory-list-2.3-small.txt",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	// Create fallback wordlist with top common recon paths
	fallbackPath := filepath.Join(os.TempDir(), "bhf_common_paths.txt")
	paths := "admin\napi\nv1\nv2\nlogin\ndashboard\nconfig\nbackup\n.git\n.env\nportal\nstaging\ntest\ndev\nswagger\ngraphql\nhealth\nstatus\nrobots.txt\nsitemap.xml\n"
	_ = os.WriteFile(fallbackPath, []byte(paths), 0644)
	return fallbackPath
}

func (t *FfufTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	wordlist := t.ensureWordlist()

	// If multiple live URLs exist, fuzz across all of them using multi-wordlist mapping
	if len(liveURLs) > 1 {
		listFile := outputFile + ".targets.txt"
		var cleanURLs []string
		for _, u := range liveURLs {
			u = strings.TrimRight(strings.TrimSpace(u), "/")
			if u != "" {
				cleanURLs = append(cleanURLs, u)
			}
		}
		if len(cleanURLs) > 1 && os.WriteFile(listFile, []byte(strings.Join(cleanURLs, "\n")), 0644) == nil {
			return []string{
				"-u", "HOST/FUZZ",
				"-w", listFile + ":HOST",
				"-w", wordlist + ":FUZZ",
				"-mc", "200,204,301,302,307,401,403",
				"-rate", "80",
				"-o", outputFile,
				"-of", "json",
			}
		}
	}

	targetURL := "https://" + CleanDomain(target.Domain)
	if len(liveURLs) > 0 {
		targetURL = strings.TrimRight(liveURLs[0], "/")
	}

	return []string{
		"-u", targetURL + "/FUZZ",
		"-w", wordlist,
		"-mc", "200,204,301,302,307,401,403",
		"-o", outputFile,
		"-of", "json",
	}
}

func (t *FfufTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *FfufTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	now := time.Now().UTC()

	// 1. Try reading the structured JSON output
	if execCtx.OutputFile != "" {
		if data, err := os.ReadFile(execCtx.OutputFile); err == nil {
			var parsed FfufJSON
			if err := json.Unmarshal(data, &parsed); err == nil {
				for _, res := range parsed.Results {
					if u, err := url.Parse(res.URL); err == nil && u.Hostname() != "" {
						assets = append(assets, models.Asset{
							ID:         uuid.New().String(),
							TargetID:   execCtx.Target.ID,
							ScanID:     execCtx.Scan.ID,
							Host:       u.Hostname(),
							URL:        res.URL,
							StatusCode: res.Status,
							CreatedAt:  now,
						})
					}
				}
			}
			_ = os.Remove(execCtx.OutputFile)
			_ = os.Remove(execCtx.OutputFile + ".targets.txt")
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

	return Result{Assets: assets}, nil
}
