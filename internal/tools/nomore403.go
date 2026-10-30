package tools

import (
	"bufio"
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

// Nomore403Tool implements the Tool interface for nomore403 HTTP bypass testing.
type Nomore403Tool struct{}

// NewNomore403Tool creates a new Nomore403Tool.
func NewNomore403Tool() *Nomore403Tool {
	return &Nomore403Tool{}
}

func (t *Nomore403Tool) Name() string {
	return "nomore403"
}

func (t *Nomore403Tool) BinaryName() string {
	return "nomore403"
}

func (t *Nomore403Tool) Category() Category {
	return CategoryFuzzing
}

func (t *Nomore403Tool) Description() string {
	return "HTTP 401/403 forbidden bypass fuzzer via header injection, verb tampering, and path mutations."
}

func (t *Nomore403Tool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "nomore403_"+scanID+".txt")
}

func (t *Nomore403Tool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	cleanRoot := CleanDomain(target.Domain)
	targetURL := "https://" + cleanRoot

	if len(liveURLs) > 0 {
		targetURL = liveURLs[0]
	} else if len(hosts) > 0 {
		targetURL = "https://" + hosts[0]
	}

	return []string{
		"-u", targetURL,
		"-max-rate", "30",
	}
}

func (t *Nomore403Tool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *Nomore403Tool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	var findings []models.Finding
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addBypassFinding := func(rawLine string) {
		rawLine = strings.TrimSpace(rawLine)
		if rawLine == "" {
			return
		}

		// Look for HTTP URL in line
		var targetURL string
		fields := strings.Fields(rawLine)
		for _, f := range fields {
			if strings.HasPrefix(f, "http://") || strings.HasPrefix(f, "https://") {
				targetURL = strings.Trim(f, "[](),")
				break
			}
		}

		if targetURL == "" {
			targetURL = "https://" + cleanRoot
		}

		parsed, err := url.Parse(targetURL)
		if err != nil || parsed.Hostname() == "" {
			return
		}

		host := strings.ToLower(parsed.Hostname())
		if !IsInScope(host, cleanRoot) || seen[rawLine] {
			return
		}
		seen[rawLine] = true

		// Check for header tampering cues
		curlCmd := fmt.Sprintf("curl -ivsL \"%s\"", targetURL)
		if strings.Contains(rawLine, "X-") {
			parts := strings.Split(rawLine, "X-")
			if len(parts) >= 2 {
				headerPart := "X-" + strings.TrimSpace(strings.Split(parts[1], " ")[0])
				curlCmd = fmt.Sprintf("curl -ivsL -H \"%s\" \"%s\"", headerPart, targetURL)
			}
		}

		findings = append(findings, models.Finding{
			ID:          uuid.New().String(),
			TargetID:    execCtx.Target.ID,
			ScanID:      execCtx.Scan.ID,
			ToolName:    t.Name(),
			Severity:    models.SeverityHigh,
			Name:        "Access Control Bypass (403/401 -> 200 OK)",
			MatchedAt:   targetURL,
			Description: fmt.Sprintf("Access restriction bypassed via header injection or path mutation: %s", rawLine),
			CurlCommand: curlCmd,
			CreatedAt:   now,
		})

		assets = append(assets, models.Asset{
			ID:            uuid.New().String(),
			TargetID:      execCtx.Target.ID,
			ScanID:        execCtx.Scan.ID,
			Host:          host,
			URL:           targetURL,
			Title:         "Bypass Succeeded: 200 OK",
			PriorityScore: 50,
			CreatedAt:     now,
		})
	}

	// 1. Read OutputFile if available
	if execCtx.OutputFile != "" {
		if f, err := os.Open(execCtx.OutputFile); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "200") || strings.Contains(line, "302") {
					addBypassFinding(line)
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
		}
	}

	// 2. Parse stdout stream lines
	for _, line := range rawOutput {
		if strings.Contains(line, "200") || strings.Contains(line, "302") {
			addBypassFinding(line)
		}
	}

	return Result{
		Assets:   assets,
		Findings: findings,
	}, nil
}
