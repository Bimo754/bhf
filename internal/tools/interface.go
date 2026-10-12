package tools

import (
	"context"
	"strings"

	"github.com/diamond/bhf/internal/models"
)

// Category organizes tools into logical methodology stages.
type Category string

const (
	CategorySubdomainEnum Category = "Subdomain Enumeration"
	CategoryHTTPProbing   Category = "HTTP Probing"
	CategoryURLDiscovery  Category = "URL Discovery"
	CategoryCrawling      Category = "Crawling & Spidering"
	CategoryVulnerability Category = "Vulnerability Scanning"
	CategoryFuzzing       Category = "Fuzzing & Content Discovery"
)

// CleanDomain normalizes a target domain by removing wildcards, protocols, and trailing slashes.
func CleanDomain(d string) string {
	d = strings.TrimSpace(strings.ToLower(d))
	d = strings.TrimPrefix(d, "*.")
	d = strings.TrimPrefix(d, ".")
	d = strings.TrimPrefix(d, "https://")
	d = strings.TrimPrefix(d, "http://")
	d = strings.TrimRight(d, "/")
	return d
}

// ExecutionContext passes target, discovered pipeline assets, and scan context to the tool execution.
type ExecutionContext struct {
	Ctx             context.Context
	Scan            *models.Scan
	Target          *models.Target
	DiscoveredHosts []string
	LiveURLs        []string
	OutputFile      string
	CustomArgs      []string
}

// Tool defines the contract for a security tool integration in BHF.
type Tool interface {
	Name() string
	BinaryName() string
	Category() Category
	Description() string
	DefaultArgs(target *models.Target, hosts []string, liveURLs []string, outputFile string) []string
	OutputFile(scanID string) string
	IsInstalled() bool
	Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error)
}
