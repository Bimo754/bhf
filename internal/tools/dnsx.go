package tools

import (
	"bufio"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// DNSxJSON represents structured JSON output from ProjectDiscovery's dnsx.
type DNSxJSON struct {
	Host  string   `json:"host"`
	A     []string `json:"a"`
	CNAME []string `json:"cname"`
}

// DNSxTool implements the Tool interface for fast DNS resolution and wildcard filtering.
type DNSxTool struct{}

// NewDNSxTool creates a new DNSxTool.
func NewDNSxTool() *DNSxTool {
	return &DNSxTool{}
}

func (t *DNSxTool) Name() string {
	return "dnsx"
}

func (t *DNSxTool) BinaryName() string {
	return "dnsx"
}

func (t *DNSxTool) Category() Category {
	return CategorySubdomainEnum
}

func (t *DNSxTool) Description() string {
	return "Fast multi-purpose DNS resolution and wildcard filtering toolkit."
}

func (t *DNSxTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "dnsx_"+scanID+".json")
}

func (t *DNSxTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	if len(hosts) > 0 {
		listFile := outputFile + ".targets.txt"
		if err := os.WriteFile(listFile, []byte(strings.Join(hosts, "\n")), 0644); err == nil {
			return []string{
				"-l", listFile,
				"-a",
				"-cname",
				"-resp",
				"-json",
				"-o", outputFile,
			}
		}
	}

	clean := CleanDomain(target.Domain)
	return []string{
		"-d", clean,
		"-a",
		"-cname",
		"-resp",
		"-json",
		"-o", outputFile,
	}
}

func (t *DNSxTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *DNSxTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addHost := func(host string, ip string) {
		host = strings.TrimSpace(strings.ToLower(host))
		if host == "" || seen[host] || !IsInScope(host, cleanRoot) {
			return
		}
		seen[host] = true

		assets = append(assets, models.Asset{
			ID:        uuid.New().String(),
			TargetID:  execCtx.Target.ID,
			ScanID:    execCtx.Scan.ID,
			Host:      host,
			IP:        ip,
			CreatedAt: now,
		})
	}

	// 1. Read JSON file if available
	if execCtx.OutputFile != "" {
		if f, err := os.Open(execCtx.OutputFile); err == nil {
			scanner := bufio.NewScanner(f)
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
					addHost(item.Host, ip)
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
			_ = os.Remove(execCtx.OutputFile + ".targets.txt")
		}
	}

	// 2. Fallback to stdout lines
	if len(assets) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "[") {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 1 && strings.Contains(parts[0], cleanRoot) {
				ip := ""
				if len(parts) >= 2 {
					ip = strings.Trim(parts[1], "[]")
				}
				addHost(parts[0], ip)
			}
		}
	}

	return Result{Assets: assets}, nil
}
