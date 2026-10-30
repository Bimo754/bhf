package tools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// HttpxJSON represents the structured JSON output from httpx.
type HttpxJSON struct {
	URL            string   `json:"url"`
	Input          string   `json:"input"`
	Host           string   `json:"host"`
	Port           string   `json:"port"`
	Title          string   `json:"title"`
	StatusCode     int      `json:"status_code"`
	Tech           []string `json:"tech"`
	ScreenshotPath string   `json:"screenshot_path"`
}

// HttpxTool implements the Tool interface for ProjectDiscovery's httpx.
type HttpxTool struct{}

// NewHttpxTool creates a new HttpxTool.
func NewHttpxTool() *HttpxTool {
	return &HttpxTool{}
}

func (t *HttpxTool) Name() string {
	return "httpx"
}

func (t *HttpxTool) BinaryName() string {
	return "httpx"
}

func (t *HttpxTool) Category() Category {
	return CategoryHTTPProbing
}

func (t *HttpxTool) Description() string {
	return "Fast multi-purpose HTTP probing toolkit for live service verification, technology detection, and visual screenshots."
}

func (t *HttpxTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "httpx_"+scanID+".json")
}

func (t *HttpxTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	dataDir := os.Getenv("BHF_DATA_DIR")
	if dataDir == "" {
		dataDir = "data"
	}
	screenshotDir := filepath.Join(dataDir, "screenshots", target.ID)
	_ = os.MkdirAll(screenshotDir, 0755)

	hasChrome := false
	if _, err := exec.LookPath("chromium"); err == nil {
		hasChrome = true
	} else if _, err := exec.LookPath("google-chrome"); err == nil {
		hasChrome = true
	} else if _, err := os.Stat("/usr/bin/chromium"); err == nil {
		hasChrome = true
	} else if _, err := os.Stat("/usr/bin/google-chrome"); err == nil {
		hasChrome = true
	}

	args := []string{"-status-code", "-title", "-tech-detect"}
	if hasChrome {
		args = append(args, "-screenshot", "-system-chrome", "-srd", screenshotDir, "-esb", "-ehb")
	}

	var targetList []string
	if len(hosts) > 0 {
		targetList = hosts
	} else if len(liveURLs) > 0 {
		targetList = liveURLs
	}

	if len(targetList) > 0 {
		listFile := outputFile + ".targets.txt"
		if err := os.WriteFile(listFile, []byte(strings.Join(targetList, "\n")), 0644); err == nil {
			args = append([]string{"-l", listFile}, args...)
			return append(args, "-o", outputFile, "-json")
		}
	}

	clean := CleanDomain(target.Domain)
	args = append([]string{"-u", clean}, args...)
	return append(args, "-o", outputFile, "-json")
}

func (t *HttpxTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *HttpxTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	now := time.Now().UTC()

	// 1. Try reading the structured JSON file if it exists
	if execCtx.OutputFile != "" {
		if f, err := os.Open(execCtx.OutputFile); err == nil {
			scanner := bufio.NewScanner(f)
			buf := make([]byte, 0, 64*1024)
			scanner.Buffer(buf, 1024*1024)

			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				var item HttpxJSON
				if err := json.Unmarshal([]byte(line), &item); err == nil && item.URL != "" {
					host := item.Input
					if host == "" {
						if parsed, err := url.Parse(item.URL); err == nil {
							host = parsed.Hostname()
						}
					}

					portNum := 0
					if item.Port != "" {
						portNum, _ = strconv.Atoi(item.Port)
					}

					techs := item.Tech
					if item.ScreenshotPath != "" {
						baseName := filepath.Base(item.ScreenshotPath)
						if baseName != "" && baseName != "." {
							screenshotURI := fmt.Sprintf("/api/screenshots/%s/%s", execCtx.Target.ID, baseName)
							techs = append(techs, "Screenshot:"+screenshotURI)
						}
					}

					assets = append(assets, models.Asset{
						ID:           uuid.New().String(),
						TargetID:     execCtx.Target.ID,
						ScanID:       execCtx.Scan.ID,
						Host:         host,
						URL:          item.URL,
						Port:         portNum,
						StatusCode:   item.StatusCode,
						Title:        item.Title,
						Technologies: techs,
						CreatedAt:    now,
					})
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
			_ = os.Remove(execCtx.OutputFile + ".targets.txt")
		}
	}

	// 2. Fallback to parsing rawOutput lines if file was empty
	if len(assets) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
				parts := strings.Fields(line)
				rawURL := parts[0]
				if parsed, err := url.Parse(rawURL); err == nil {
					assets = append(assets, models.Asset{
						ID:        uuid.New().String(),
						TargetID:  execCtx.Target.ID,
						ScanID:    execCtx.Scan.ID,
						Host:      parsed.Hostname(),
						URL:       rawURL,
						CreatedAt: now,
					})
				}
			}
		}
	}

	return Result{Assets: assets}, nil
}
