package tools

import (
	"bufio"
	"encoding/json"
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

// TlsxJSON represents structured JSON output from ProjectDiscovery's tlsx.
type TlsxJSON struct {
	Host        string   `json:"host"`
	IP          string   `json:"ip"`
	Port        string   `json:"port"`
	ProbeStatus bool     `json:"probe_status"`
	SubjectAN   []string `json:"subject_an"`
	SubjectCN   string   `json:"subject_cn"`
	Jarm        string   `json:"jarm"`
	TLSVersion  string   `json:"tls_version"`
}

// TlsxTool implements the Tool interface for ProjectDiscovery's tlsx.
type TlsxTool struct{}

// NewTlsxTool creates a new TlsxTool.
func NewTlsxTool() *TlsxTool {
	return &TlsxTool{}
}

func (t *TlsxTool) Name() string {
	return "tlsx"
}

func (t *TlsxTool) BinaryName() string {
	return "tlsx"
}

func (t *TlsxTool) Category() Category {
	return CategoryHTTPProbing
}

func (t *TlsxTool) Description() string {
	return "Fast TLS/SSL grabber for certificate intelligence, SAN extraction, and JARM fingerprinting."
}

func (t *TlsxTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "tlsx_"+scanID+".json")
}

func (t *TlsxTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
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
			"-san",
			"-cn",
			"-jarm",
			"-json",
			"-o", outputFile,
			"-c", "40",
			"-silent",
		}
	}

	return []string{
		"-u", cleanRoot,
		"-san",
		"-cn",
		"-jarm",
		"-json",
		"-o", outputFile,
		"-c", "40",
		"-silent",
	}
}

func (t *TlsxTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *TlsxTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addHost := func(host string, ip string, portStr string, techs []string) {
		host = strings.TrimSpace(strings.ToLower(host))
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
		if host == "" || seen[host] || !IsInScope(host, cleanRoot) {
			return
		}
		seen[host] = true

		port := 443
		if portStr != "" {
			if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
				port = p
			}
		}

		assets = append(assets, models.Asset{
			ID:           uuid.New().String(),
			TargetID:     execCtx.Target.ID,
			ScanID:       execCtx.Scan.ID,
			Host:         host,
			IP:           ip,
			Port:         port,
			Technologies: techs,
			CreatedAt:    now,
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

				var item TlsxJSON
				if err := json.Unmarshal([]byte(line), &item); err == nil {
					var techs []string
					if item.TLSVersion != "" {
						techs = append(techs, "TLS:"+item.TLSVersion)
					}
					if item.Jarm != "" {
						techs = append(techs, "JARM:"+item.Jarm[:8])
					}

					baseHost := item.Host
					if parsed, err := url.Parse(item.Host); err == nil && parsed.Hostname() != "" {
						baseHost = parsed.Hostname()
					}

					addHost(baseHost, item.IP, item.Port, techs)

					if item.SubjectCN != "" {
						addHost(item.SubjectCN, item.IP, item.Port, techs)
					}
					for _, san := range item.SubjectAN {
						addHost(san, item.IP, item.Port, techs)
					}
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
		}
		_ = os.Remove(execCtx.OutputFile + ".targets.txt")
	}

	// 2. Parse stdout lines fallback
	if len(assets) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "[") {
				continue
			}

			var item TlsxJSON
			if strings.HasPrefix(line, "{") && json.Unmarshal([]byte(line), &item) == nil {
				baseHost := item.Host
				if idx := strings.Index(baseHost, ":"); idx != -1 {
					baseHost = baseHost[:idx]
				}
				addHost(baseHost, item.IP, item.Port, nil)
				for _, san := range item.SubjectAN {
					addHost(san, item.IP, item.Port, nil)
				}
				continue
			}

			parts := strings.Fields(line)
			if len(parts) >= 1 && strings.Contains(parts[0], cleanRoot) {
				host := parts[0]
				ip := ""
				if len(parts) >= 2 {
					ip = strings.Trim(parts[1], "[]")
				}
				addHost(host, ip, "443", nil)
			}
		}
	}

	return Result{Assets: assets}, nil
}
