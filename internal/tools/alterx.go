package tools

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// AlterxTool implements the Tool interface for ProjectDiscovery's alterx.
type AlterxTool struct{}

// NewAlterxTool creates a new AlterxTool.
func NewAlterxTool() *AlterxTool {
	return &AlterxTool{}
}

func (t *AlterxTool) Name() string {
	return "alterx"
}

func (t *AlterxTool) BinaryName() string {
	return "alterx"
}

func (t *AlterxTool) Category() Category {
	return CategorySubdomainEnum
}

func (t *AlterxTool) Description() string {
	return "Subdomain wordlist permutation and alteration generator."
}

func (t *AlterxTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "alterx_"+scanID+".txt")
}

func (t *AlterxTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
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
			"-l", listFile,
			"-en",
			"-silent",
		}
	}

	return []string{
		"-enrich",
		"-silent",
	}
}

func (t *AlterxTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *AlterxTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addHost := func(host string) {
		host = strings.TrimSpace(strings.ToLower(host))
		if idx := strings.Index(host, ":"); idx != -1 {
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
			CreatedAt: now,
		})
	}

	// 1. Read from output file
	if execCtx.OutputFile != "" {
		if f, err := os.Open(execCtx.OutputFile); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" && !strings.HasPrefix(line, "[") {
					addHost(line)
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
		}
		_ = os.Remove(execCtx.OutputFile + ".targets.txt")
	}

	// 2. Fallback to raw output lines
	if len(assets) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "[") || strings.Contains(line, " ") {
				continue
			}
			if strings.Contains(line, cleanRoot) {
				addHost(line)
			}
		}
	}

	return Result{Assets: assets}, nil
}
