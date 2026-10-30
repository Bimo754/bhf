package tools

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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
	return "Subdomain wordlist permutation and alteration generator with DNS resolution verification."
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
	var candidates []string
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addCandidate := func(host string) {
		host = strings.TrimSpace(strings.ToLower(host))
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
		if host == "" || seen[host] || !IsInScope(host, cleanRoot) {
			return
		}
		seen[host] = true
		candidates = append(candidates, host)
	}

	// 1. Read from output file
	if execCtx.OutputFile != "" {
		if f, err := os.Open(execCtx.OutputFile); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" && !strings.HasPrefix(line, "[") {
					addCandidate(line)
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
		}
		_ = os.Remove(execCtx.OutputFile + ".targets.txt")
	}

	// 2. Fallback to raw output lines
	if len(candidates) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "[") || strings.Contains(line, " ") {
				continue
			}
			if strings.Contains(line, cleanRoot) {
				addCandidate(line)
			}
		}
	}

	if len(candidates) == 0 {
		return Result{}, nil
	}

	var assets []models.Asset

	// 3. DNS Resolution Verification Pipeline (FEAT-014):
	// If dnsx binary is available in PATH, pipe permutations to dnsx to only save verified live DNS records.
	if dnsxBinary, err := exec.LookPath("dnsx"); err == nil && len(candidates) > 0 {
		ctx := execCtx.Ctx
		if ctx == nil {
			ctx = context.Background()
		}
		cmd := exec.CommandContext(ctx, dnsxBinary, "-silent", "-a", "-resp", "-json")
		cmd.Stdin = strings.NewReader(strings.Join(candidates, "\n"))
		out, err := cmd.Output()
		if err == nil && len(out) > 0 {
			scanner := bufio.NewScanner(bytes.NewReader(out))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}
				var item DNSxJSON
				if err := json.Unmarshal([]byte(line), &item); err == nil && item.Host != "" {
					ip := ""
					if len(item.A) > 0 {
						ip = item.A[0]
					}
					// Only record if valid A or CNAME returned
					if len(item.A) > 0 || len(item.CNAME) > 0 {
						assets = append(assets, models.Asset{
							ID:        uuid.New().String(),
							TargetID:  execCtx.Target.ID,
							ScanID:    execCtx.Scan.ID,
							Host:      item.Host,
							IP:        ip,
							CreatedAt: now,
						})
					}
				}
			}
			return Result{Assets: assets}, nil
		}
	}

	// Fallback when dnsx is not installed (e.g. testing or minimal runner environment):
	for _, c := range candidates {
		assets = append(assets, models.Asset{
			ID:        uuid.New().String(),
			TargetID:  execCtx.Target.ID,
			ScanID:    execCtx.Scan.ID,
			Host:      c,
			CreatedAt: now,
		})
	}

	return Result{Assets: assets}, nil
}
