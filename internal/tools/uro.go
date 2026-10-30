package tools

import (
	"bufio"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// UroTool implements the Tool interface for uro URL decluttering and deduplication.
type UroTool struct{}

// NewUroTool creates a new UroTool.
func NewUroTool() *UroTool {
	return &UroTool{}
}

func (t *UroTool) Name() string {
	return "uro"
}

func (t *UroTool) BinaryName() string {
	if _, err := exec.LookPath("uro"); err == nil {
		return "uro"
	}
	return "uv"
}

func (t *UroTool) Category() Category {
	return CategoryURLDiscovery
}

func (t *UroTool) Description() string {
	return "URL decluttering and deduplication filter for cleaning crawler archives."
}

func (t *UroTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "uro_"+scanID+".txt")
}

func (t *UroTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	var inputURLs []string
	cleanRoot := CleanDomain(target.Domain)

	if len(liveURLs) > 0 {
		for _, u := range liveURLs {
			u = strings.TrimSpace(u)
			if u != "" {
				inputURLs = append(inputURLs, u)
			}
		}
	} else if len(hosts) > 0 {
		for _, h := range hosts {
			h = strings.TrimSpace(h)
			if h != "" && IsInScope(h, cleanRoot) {
				inputURLs = append(inputURLs, "https://"+h+"/")
			}
		}
	}

	if len(inputURLs) == 0 {
		inputURLs = append(inputURLs, "https://"+cleanRoot+"/")
	}

	listFile := outputFile + ".targets.txt"
	_ = os.WriteFile(listFile, []byte(strings.Join(inputURLs, "\n")), 0644)

	baseArgs := []string{
		"-i", listFile,
		"-o", outputFile,
	}

	if t.BinaryName() == "uv" {
		return append([]string{"run", "uro"}, baseArgs...)
	}

	return baseArgs
}

func (t *UroTool) IsInstalled() bool {
	_, err1 := exec.LookPath("uro")
	_, err2 := exec.LookPath("uv")
	return err1 == nil || err2 == nil
}

func (t *UroTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addURL := func(rawURL string) {
		rawURL = strings.TrimSpace(rawURL)
		if rawURL == "" || seen[rawURL] {
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
		seen[rawURL] = true

		assets = append(assets, models.Asset{
			ID:        uuid.New().String(),
			TargetID:  execCtx.Target.ID,
			ScanID:    execCtx.Scan.ID,
			Host:      host,
			URL:       rawURL,
			CreatedAt: now,
		})
	}

	// 1. Read clean output file
	if execCtx.OutputFile != "" {
		if f, err := os.Open(execCtx.OutputFile); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					addURL(line)
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
		}
		_ = os.Remove(execCtx.OutputFile + ".targets.txt")
	}

	// 2. Fallback to stdout lines
	if len(assets) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if line != "" && strings.HasPrefix(line, "http") {
				addURL(line)
			}
		}
	}

	return Result{Assets: assets}, nil
}
