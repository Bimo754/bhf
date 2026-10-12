package tools

import (
	"bufio"
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/diamond/bhf/internal/models"
	"github.com/google/uuid"
)

// NucleiJSON represents a finding output from nuclei -jsonl.
type NucleiJSON struct {
	TemplateID string `json:"template-id"`
	Info       struct {
		Name     string   `json:"name"`
		Severity string   `json:"severity"`
		Tags     []string `json:"tags"`
	} `json:"info"`
	Host    string `json:"host"`
	Matched string `json:"matched-at"`
	IP      string `json:"ip"`
}

// NucleiTool implements the Tool interface for ProjectDiscovery's Nuclei.
type NucleiTool struct{}

// NewNucleiTool creates a new NucleiTool.
func NewNucleiTool() *NucleiTool {
	return &NucleiTool{}
}

func (t *NucleiTool) Name() string {
	return "nuclei"
}

func (t *NucleiTool) BinaryName() string {
	return "nuclei"
}

func (t *NucleiTool) Category() Category {
	return CategoryVulnerability
}

func (t *NucleiTool) Description() string {
	return "Fast template-based vulnerability and exposure scanner."
}

func (t *NucleiTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "nuclei_"+scanID+".jsonl")
}

func (t *NucleiTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	// If live URLs exist, scan them! Otherwise scan discovered hosts, or fallback to root domain.
	rawTargets := liveURLs
	if len(rawTargets) == 0 {
		rawTargets = hosts
	}

	cleanRoot := CleanDomain(target.Domain)
	var targets []string
	for _, item := range rawTargets {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if u, err := url.Parse(item); err == nil && u.Hostname() != "" {
			if IsInScope(u.Hostname(), cleanRoot) {
				targets = append(targets, item)
			}
		} else if IsInScope(item, cleanRoot) {
			targets = append(targets, item)
		}
	}

	if len(targets) > 0 {
		listFile := outputFile + ".targets.txt"
		if err := os.WriteFile(listFile, []byte(strings.Join(targets, "\n")), 0644); err == nil {
			return []string{
				"-l", listFile,
				"-s", "info,low,medium,high,critical",
				"-c", "25",
				"-rl", "150",
				"-stats",
				"-si", "10",
				"-jle", outputFile,
			}
		}
	}

	return []string{
		"-u", cleanRoot,
		"-s", "info,low,medium,high,critical",
		"-c", "25",
		"-rl", "150",
		"-stats",
		"-si", "10",
		"-jle", outputFile,
	}
}

func (t *NucleiTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *NucleiTool) Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error) {
	var assets []models.Asset
	now := time.Now().UTC()

	// 1. Try reading the structured JSON file
	if execCtx.OutputFile != "" {
		if f, err := os.Open(execCtx.OutputFile); err == nil {
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					continue
				}

				var item NucleiJSON
				if err := json.Unmarshal([]byte(line), &item); err == nil && item.Host != "" {
					host := item.Host
					if parsed, err := url.Parse(item.Host); err == nil && parsed.Hostname() != "" {
						host = parsed.Hostname()
					}

					assets = append(assets, models.Asset{
						ID:           uuid.New().String(),
						TargetID:     execCtx.Target.ID,
						ScanID:       execCtx.Scan.ID,
						Host:         host,
						URL:          item.Matched,
						IP:           item.IP,
						Title:        item.Info.Name + " [" + strings.ToUpper(item.Info.Severity) + "]",
						Technologies: item.Info.Tags,
						CreatedAt:    now,
					})
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
			_ = os.Remove(execCtx.OutputFile + ".targets.txt")
		}
	}

	// 2. Fallback to scanning rawOutput for matches
	if len(assets) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "[") && strings.Contains(line, "http") {
				parts := strings.Fields(line)
				for _, part := range parts {
					if strings.HasPrefix(part, "http") {
						if parsed, err := url.Parse(part); err == nil {
							assets = append(assets, models.Asset{
								ID:        uuid.New().String(),
								TargetID:  execCtx.Target.ID,
								ScanID:    execCtx.Scan.ID,
								Host:      parsed.Hostname(),
								URL:       part,
								Title:     line,
								CreatedAt: now,
							})
						}
					}
				}
			}
		}
	}

	return assets, nil
}
