package tools

import (
	"bufio"
	"encoding/json"
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

// S3ScannerJSON represents a single result item from s3scanner JSON output.
type S3ScannerJSON struct {
	BucketName string `json:"bucket_name"`
	Exists     bool   `json:"exists"`
	AllUsers   string `json:"all_users"`
	AuthUsers  string `json:"auth_users"`
	Provider   string `json:"provider"`
	BucketURL  string `json:"bucket_url"`
}

// S3ScannerTool implements the Tool interface for s3scanner cloud storage reconnaissance.
type S3ScannerTool struct{}

// NewS3ScannerTool creates a new S3ScannerTool.
func NewS3ScannerTool() *S3ScannerTool {
	return &S3ScannerTool{}
}

func (t *S3ScannerTool) Name() string {
	return "s3scanner"
}

func (t *S3ScannerTool) BinaryName() string {
	return "s3scanner"
}

func (t *S3ScannerTool) Category() Category {
	return CategoryVulnerability
}

func (t *S3ScannerTool) Description() string {
	return "Cloud storage bucket reconnaissance scanner for public Amazon S3, Google Cloud, and Azure Blobs."
}

func (t *S3ScannerTool) OutputFile(scanID string) string {
	return filepath.Join(os.TempDir(), "s3scanner_"+scanID+".json")
}

func (t *S3ScannerTool) DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string {
	cleanRoot := CleanDomain(target.Domain)
	rootPrefix := strings.Split(cleanRoot, ".")[0]

	candidates := []string{
		cleanRoot,
		strings.ReplaceAll(cleanRoot, ".", "-"),
		rootPrefix,
		rootPrefix + "-dev",
		rootPrefix + "-staging",
		rootPrefix + "-backup",
		rootPrefix + "-assets",
		rootPrefix + "-data",
		rootPrefix + "-media",
	}

	for _, h := range hosts {
		h = strings.TrimSpace(h)
		if h != "" && IsInScope(h, cleanRoot) {
			prefix := strings.Split(h, ".")[0]
			if prefix != "" && prefix != rootPrefix {
				candidates = append(candidates, prefix+"-"+rootPrefix)
			}
		}
	}

	listFile := outputFile + ".buckets.txt"
	if err := os.WriteFile(listFile, []byte(strings.Join(candidates, "\n")), 0644); err == nil {
		return []string{
			"-bucket-file", listFile,
			"-json",
		}
	}

	return []string{
		"-bucket", rootPrefix,
		"-json",
	}
}

func (t *S3ScannerTool) IsInstalled() bool {
	_, err := exec.LookPath(t.BinaryName())
	return err == nil
}

func (t *S3ScannerTool) Run(execCtx ExecutionContext, rawOutput []string) (Result, error) {
	var assets []models.Asset
	var findings []models.Finding
	seen := make(map[string]bool)
	now := time.Now().UTC()

	addBucketFinding := func(name string, provider string, perm string, bURL string) {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			return
		}
		seen[name] = true

		if provider == "" {
			provider = "Amazon S3"
		}
		if bURL == "" {
			bURL = fmt.Sprintf("https://%s.s3.amazonaws.com", name)
		}

		sev := models.SeverityMedium
		if strings.Contains(strings.ToLower(perm), "read") || strings.Contains(strings.ToLower(perm), "write") {
			sev = models.SeverityHigh
		}

		curlCmd := fmt.Sprintf("curl -ivsL \"%s\"", bURL)
		findings = append(findings, models.Finding{
			ID:          uuid.New().String(),
			TargetID:    execCtx.Target.ID,
			ScanID:      execCtx.Scan.ID,
			ToolName:    t.Name(),
			Severity:    sev,
			Name:        fmt.Sprintf("Exposed Cloud Storage Bucket [%s]", name),
			MatchedAt:   bURL,
			Description: fmt.Sprintf("Public cloud storage bucket discovered (%s) with permissions: %s", provider, perm),
			CurlCommand: curlCmd,
			CreatedAt:   now,
		})

		parsed, err := url.Parse(bURL)
		host := name + ".s3.amazonaws.com"
		if err == nil && parsed.Hostname() != "" {
			host = parsed.Hostname()
		}

		assets = append(assets, models.Asset{
			ID:            uuid.New().String(),
			TargetID:      execCtx.Target.ID,
			ScanID:        execCtx.Scan.ID,
			Host:          host,
			URL:           bURL,
			Title:         fmt.Sprintf("Cloud Bucket: %s (%s)", name, provider),
			Technologies:  []string{"Cloud Storage:" + provider},
			PriorityScore: 45,
			CreatedAt:     now,
		})
	}

	// 1. Read JSON file if available
	if execCtx.OutputFile != "" {
		if data, err := os.ReadFile(execCtx.OutputFile); err == nil && len(data) > 0 {
			var items []S3ScannerJSON
			if err := json.Unmarshal(data, &items); err == nil {
				for _, item := range items {
					if item.Exists {
						perm := "exists"
						if item.AllUsers != "" {
							perm = "all_users:" + item.AllUsers
						}
						addBucketFinding(item.BucketName, item.Provider, perm, item.BucketURL)
					}
				}
			} else {
				scanner := bufio.NewScanner(strings.NewReader(string(data)))
				for scanner.Scan() {
					var item S3ScannerJSON
					if json.Unmarshal([]byte(scanner.Text()), &item) == nil && item.Exists {
						perm := "exists"
						if item.AllUsers != "" {
							perm = "all_users:" + item.AllUsers
						}
						addBucketFinding(item.BucketName, item.Provider, perm, item.BucketURL)
					}
				}
			}
			_ = os.Remove(execCtx.OutputFile)
		}
		_ = os.Remove(execCtx.OutputFile + ".buckets.txt")
	}

	// 2. Parse stdout stream lines
	for _, line := range rawOutput {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") {
			var logEntry struct {
				Bucket struct {
					Name              string `json:"name"`
					Exists            int    `json:"exists"`
					Provider          string `json:"provider"`
					PermAllUsersRead  int    `json:"perm_all_users_read"`
					PermAllUsersWrite int    `json:"perm_all_users_write"`
				} `json:"bucket"`
			}
			if err := json.Unmarshal([]byte(line), &logEntry); err == nil && logEntry.Bucket.Name != "" {
				if logEntry.Bucket.Exists == 1 || logEntry.Bucket.PermAllUsersRead == 1 || logEntry.Bucket.PermAllUsersWrite == 1 {
					perm := "exists"
					if logEntry.Bucket.PermAllUsersRead == 1 {
						perm = "read"
					}
					if logEntry.Bucket.PermAllUsersWrite == 1 {
						perm = "write"
					}
					provider := logEntry.Bucket.Provider
					if provider == "" {
						provider = "aws"
					}
					addBucketFinding(logEntry.Bucket.Name, strings.ToUpper(provider), perm, "")
				}
				continue
			}
		}

		if strings.Contains(line, "|") && (strings.Contains(line, "exists") || strings.Contains(line, "read")) {
			parts := strings.Split(line, "|")
			if len(parts) >= 2 {
				name := strings.TrimSpace(parts[0])
				perm := strings.TrimSpace(strings.Join(parts[1:], " | "))
				addBucketFinding(name, "AWS S3", perm, "")
			}
		} else if strings.Contains(line, "[FOUND]") || strings.Contains(line, "[OPEN]") {
			parts := strings.Fields(line)
			for _, p := range parts {
				if !strings.HasPrefix(p, "[") && len(p) > 2 {
					addBucketFinding(p, "Cloud Storage", "open", "")
					break
				}
			}
		}
	}

	return Result{
		Assets:   assets,
		Findings: findings,
	}, nil
}
