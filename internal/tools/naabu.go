package tools

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// NaabuJSON represents structured JSON output from ProjectDiscovery's naabu.
type NaabuJSON struct {
	Host string `json:"host"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
}

// NaabuTool implements the Tool interface for fast port scanning.
type NaabuTool struct{}

// NewNaabuTool creates a new NaabuTool.
func NewNaabuTool() *NaabuTool {
	return &NaabuTool{}
}

func (t *NaabuTool) Name() string {
	return "naabu"
}

func (t *NaabuTool) BinaryName() string {
	return "naabu"
}

func (t *NaabuTool) Category() Category {
	return CategoryPortScanning
}

func (t *NaabuTool) Description() string {
	return "Fast port scanner for discovering open network ports and exposed services."
}

func (t *NaabuTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "naabu_"+scanID+".json")
}

func (t *NaabuTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	if len(hosts) > 0 {
		listFile := outputFile + ".targets.txt"
		if err := os.WriteFile(listFile, []byte(strings.Join(hosts, "\n")), 0644); err == nil {
			return []string{
				"-l", listFile,
				"-tp", "100",
				"-s", "c",
				"-c", "50",
				"-json",
				"-o", outputFile,
			}
		}
	}

	clean := CleanDomain(target.Domain)
	return []string{
		"-host", clean,
		"-tp", "100",
		"-s", "c",
		"-c", "50",
		"-json",
		"-o", outputFile,
	}
}

func (t *NaabuTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *NaabuTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addPort := func(host string, ip string, port int) {
		host = strings.TrimSpace(strings.ToLower(host))
		key := host + ":" + strconv.Itoa(port)
		if host == "" || seen[key] || !IsInScope(host, cleanRoot) {
			return
		}
		seen[key] = true

		assets = append(assets, models.Asset{
			ID:        uuid.New().String(),
			TargetID:  execCtx.Target.ID,
			ScanID:    execCtx.Scan.ID,
			Host:      host,
			IP:        ip,
			Port:      port,
			CreatedAt: now,
		})
	}

	// 1. Read structured JSON file if generated
	if execCtx.OutputFile != "" {
		if f, err := os.Open(execCtx.OutputFile); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				var item NaabuJSON
				if err := json.Unmarshal([]byte(line), &item); err == nil && item.Port > 0 {
					host := item.Host
					if host == "" {
						host = item.IP
					}
					addPort(host, item.IP, item.Port)
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
			_ = os.Remove(execCtx.OutputFile + ".targets.txt")
		}
	}

	// 2. Fallback to parsing stdout lines
	if len(assets) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "[") {
				continue
			}

			if host, portStr, err := net.SplitHostPort(line); err == nil {
				if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
					addPort(host, "", port)
				}
			}
		}
	}

	return Result{Assets: assets}, nil
}
