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
		Name        string   `json:"name"`
		Severity    string   `json:"severity"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	} `json:"info"`
	Host        string   `json:"host"`
	Matched     string   `json:"matched-at"`
	IP          string   `json:"ip"`
	CurlCommand string   `json:"curl-command"`
	Extracted   []string `json:"extracted-results"`
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

	baseArgs := []string{
		"-s", "low,medium,high,critical",
		"-tags", "cve,exposure,misconfig,takeover,vuln",
		"-c", "50",
		"-rl", "300",
		"-timeout", "5",
		"-retries", "1",
		"-mhe", "3",
		"-stats",
		"-si", "10",
		"-jle", outputFile,
	}

	if len(targets) > 0 {
		listFile := outputFile + ".targets.txt"
		if err := os.WriteFile(listFile, []byte(strings.Join(targets, "\n")), 0644); err == nil {
			return append([]string{"-l", listFile}, baseArgs...)
		}
	}

	return append([]string{"-u", cleanRoot}, baseArgs...)
}

func (t *NucleiTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *NucleiTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var findings []models.Finding
	var assets []models.Asset
	now := time.Now().UTC()

	// 1. Try reading the structured JSONL file
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

					name := item.Info.Name
					if name == "" {
						name = item.TemplateID
					}

					sev := models.Severity(strings.ToLower(item.Info.Severity))
					if sev == "" {
						sev = models.SeverityInfo
					}

					findings = append(findings, models.Finding{
						ID:               uuid.New().String(),
						TargetID:         execCtx.Target.ID,
						ScanID:           execCtx.Scan.ID,
						ToolName:         t.Name(),
						Severity:         sev,
						Name:             name,
						MatchedAt:        item.Matched,
						Description:      item.Info.Description,
						ExtractedResults: item.Extracted,
						CurlCommand:      item.CurlCommand,
						CreatedAt:        now,
					})

					// Record asset if host is known
					if host != "" {
						assets = append(assets, models.Asset{
							ID:           uuid.New().String(),
							TargetID:     execCtx.Target.ID,
							ScanID:       execCtx.Scan.ID,
							Host:         host,
							URL:          item.Matched,
							IP:           item.IP,
							Title:        name + " [" + strings.ToUpper(string(sev)) + "]",
							Technologies: item.Info.Tags,
							CreatedAt:    now,
						})
					}
				}
			}
			_ = f.Close()
			_ = os.Remove(execCtx.OutputFile)
			_ = os.Remove(execCtx.OutputFile + ".targets.txt")
		}
	}

	// 2. Fallback to scanning rawOutput lines for matches
	if len(findings) == 0 {
		for _, line := range rawOutput {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "[") && strings.Contains(line, "http") {
				parts := strings.Fields(line)
				for _, part := range parts {
					if strings.HasPrefix(part, "http") {
						findings = append(findings, models.Finding{
							ID:        uuid.New().String(),
							TargetID:  execCtx.Target.ID,
							ScanID:    execCtx.Scan.ID,
							ToolName:  t.Name(),
							Severity:  models.SeverityInfo,
							Name:      line,
							MatchedAt: part,
							CreatedAt: now,
						})
					}
				}
			}
		}
	}

	return Result{
		Assets:   assets,
		Findings: findings,
	}, nil
}
