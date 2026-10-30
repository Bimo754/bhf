package tools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// AsnmapJSON represents a single JSON output item from asnmap.
type AsnmapJSON struct {
	FirstIP   string `json:"first_ip"`
	LastIP    string `json:"last_ip"`
	Input     string `json:"input"`
	ASN       any    `json:"asn"`
	ASName    string `json:"as_name"`
	ASCountry string `json:"as_country"`
	CIDR      string `json:"cidr"`
}

// AsnmapTool implements the Tool interface for ProjectDiscovery's asnmap.
type AsnmapTool struct{}

// NewAsnmapTool creates a new AsnmapTool.
func NewAsnmapTool() *AsnmapTool {
	return &AsnmapTool{}
}

func (t *AsnmapTool) Name() string {
	return "asnmap"
}

func (t *AsnmapTool) BinaryName() string {
	return "asnmap"
}

func (t *AsnmapTool) Category() Category {
	return CategorySubdomainEnum
}

func (t *AsnmapTool) Description() string {
	return "Network range, ASN, and CIDR mapping tool."
}

func (t *AsnmapTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "asnmap_"+scanID+".json")
}

func (t *AsnmapTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	cleanRoot := CleanDomain(target.Domain)
	return []string{
		"-d", cleanRoot,
		"-json",
		"-o", outputFile,
		"-silent",
	}
}

func (t *AsnmapTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *AsnmapTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	seen := make(map[string]bool)
	now := time.Now().UTC()
	cleanRoot := CleanDomain(execCtx.Target.Domain)

	addNetworkAsset := func(ipCIDR string, asnStr string, asName string, country string) {
		key := ipCIDR + ":" + asnStr
		if ipCIDR == "" || seen[key] {
			return
		}
		seen[key] = true

		title := "Network Allocation"
		if asnStr != "" && asName != "" {
			title = fmt.Sprintf("ASN %s - %s", asnStr, asName)
			if country != "" {
				title += " (" + country + ")"
			}
		} else if asnStr != "" {
			title = fmt.Sprintf("ASN %s", asnStr)
		}

		var techs []string
		if asnStr != "" {
			techs = append(techs, "ASN:"+asnStr)
		}
		if asName != "" {
			techs = append(techs, asName)
		}

		assets = append(assets, models.Asset{
			ID:           uuid.New().String(),
			TargetID:     execCtx.Target.ID,
			ScanID:       execCtx.Scan.ID,
			Host:         cleanRoot,
			IP:           ipCIDR,
			Title:        title,
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

				var item AsnmapJSON
				if err := json.Unmarshal([]byte(line), &item); err == nil {
					asnStr := ""
					if item.ASN != nil {
						asnStr = fmt.Sprintf("%v", item.ASN)
					}
					ipTarget := item.CIDR
					if ipTarget == "" {
						ipTarget = item.FirstIP
					}
					addNetworkAsset(ipTarget, asnStr, item.ASName, item.ASCountry)
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
		}
	}

	// 2. Parse stdout lines fallback
	if len(assets) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "[") {
				continue
			}

			// Check for JSON line in stdout
			var item AsnmapJSON
			if strings.HasPrefix(line, "{") && json.Unmarshal([]byte(line), &item) == nil {
				asnStr := ""
				if item.ASN != nil {
					asnStr = fmt.Sprintf("%v", item.ASN)
				}
				ipTarget := item.CIDR
				if ipTarget == "" {
					ipTarget = item.FirstIP
				}
				addNetworkAsset(ipTarget, asnStr, item.ASName, item.ASCountry)
				continue
			}

			// Plain text line (e.g. CIDR or ASN)
			if strings.Contains(line, "/") {
				addNetworkAsset(line, "", "", "")
			} else if strings.HasPrefix(strings.ToUpper(line), "AS") {
				addNetworkAsset("", line, "", "")
			}
		}
	}

	return Result{Assets: assets}, nil
}
