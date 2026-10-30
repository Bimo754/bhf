package tools

import (
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

// GowitnessTool implements the Tool interface for gowitness web screenshotting.
type GowitnessTool struct{}

// NewGowitnessTool creates a new GowitnessTool.
func NewGowitnessTool() *GowitnessTool {
	return &GowitnessTool{}
}

func (t *GowitnessTool) Name() string {
	return "gowitness"
}

func (t *GowitnessTool) BinaryName() string {
	return "gowitness"
}

func (t *GowitnessTool) Category() Category {
	return CategoryHTTPProbing
}

func (t *GowitnessTool) Description() string {
	return "Headless browser screenshot utility for visual reconnaissance triage."
}

func (t *GowitnessTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "gowitness_"+scanID+".json")
}

func (t *GowitnessTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
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

	dataDir := os.Getenv("BHF_DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}
	screenshotDir := filepath.Join(dataDir, "screenshots", target.ID)
	_ = os.MkdirAll(screenshotDir, 0755)

	var chromeArgs []string
	if p, err := exec.LookPath("chromium"); err == nil {
		chromeArgs = []string{"--chrome-path", p}
	} else if p, err := exec.LookPath("google-chrome"); err == nil {
		chromeArgs = []string{"--chrome-path", p}
	} else if _, err := os.Stat("/usr/bin/chromium"); err == nil {
		chromeArgs = []string{"--chrome-path", "/usr/bin/chromium"}
	}

	listFile := outputFile + ".targets.txt"
	if err := os.WriteFile(listFile, []byte(strings.Join(targets, "\n")), 0644); err == nil {
		base := []string{
			"scan",
			"file",
			"-f", listFile,
			"--screenshot-path", screenshotDir,
			"--write-stdout",
			"--delay", "2",
		}
		return append(base, chromeArgs...)
	}

	base := []string{
		"scan",
		"single",
		"-u", targets[0],
		"--screenshot-path", screenshotDir,
		"--write-stdout",
		"--delay", "2",
	}
	return append(base, chromeArgs...)
}

func (t *GowitnessTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *GowitnessTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	dataDir := os.Getenv("BHF_DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}
	screenshotDir := filepath.Join(dataDir, "screenshots", execCtx.Target.ID)
	var screenshotFiles []string
	if entries, err := os.ReadDir(screenshotDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && (strings.HasSuffix(e.Name(), ".jpeg") || strings.HasSuffix(e.Name(), ".jpg") || strings.HasSuffix(e.Name(), ".png")) {
				screenshotFiles = append(screenshotFiles, e.Name())
			}
		}
	}

	findScreenshotURL := func(host string) string {
		for _, f := range screenshotFiles {
			normalized := strings.ReplaceAll(strings.ToLower(f), "-", ".")
			if strings.Contains(normalized, host) {
				return fmt.Sprintf("/api/screenshots/%s/%s", execCtx.Target.ID, f)
			}
		}
		if len(screenshotFiles) == 1 {
			return fmt.Sprintf("/api/screenshots/%s/%s", execCtx.Target.ID, screenshotFiles[0])
		}
		return ""
	}

	addScreenshotAsset := func(targetURL string) {
		targetURL = strings.TrimSpace(targetURL)
		if targetURL == "" || seen[targetURL] {
			return
		}

		parsed, err := url.Parse(targetURL)
		if err != nil || parsed.Hostname() == "" {
			return
		}

		host := strings.ToLower(parsed.Hostname())
		if !IsInScope(host, cleanRoot) {
			return
		}
		seen[targetURL] = true

		techs := []string{"Visual Screenshot"}
		if sURL := findScreenshotURL(host); sURL != "" {
			techs = append(techs, "Screenshot:"+sURL)
		}

		assets = append(assets, models.Asset{
			ID:           uuid.New().String(),
			TargetID:     execCtx.Target.ID,
			ScanID:       execCtx.Scan.ID,
			Host:         host,
			URL:          targetURL,
			Title:        "Visual Screenshot Captured",
			Technologies: techs,
			CreatedAt:    now,
		})
	}

	// Clean up temporary targets file
	if execCtx.OutputFile != "" {
		_ = os.Remove(execCtx.OutputFile + ".targets.txt")
		_ = os.Remove(execCtx.OutputFile)
	}

	// Parse stdout lines for screenshot capture confirmations
	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "http://") || strings.Contains(line, "https://") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
					addScreenshotAsset(strings.Trim(p, "[](),\"'"))
				}
			}
		}
	}

	// If no assets parsed from stdout, map live targets from execution context
	if len(assets) == 0 && len(execCtx.LiveURLs) > 0 {
		for _, u := range execCtx.LiveURLs {
			addScreenshotAsset(u)
		}
	}

	return Result{Assets: assets}, nil
}
