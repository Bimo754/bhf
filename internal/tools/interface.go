package tools

import (
	"context"

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

// ExecutionContext passes target and scan context to the tool execution.
type ExecutionContext struct {
	Ctx      context.Context
	Scan     *models.Scan
	Target   *models.Target
	CustomArgs []string
}

// Tool defines the contract for a security tool integration in BHF.
type Tool interface {
	Name() string
	BinaryName() string
	Category() Category
	Description() string
	DefaultArgs(targetDomain string) []string
	IsInstalled() bool
	Run(execCtx ExecutionContext, rawOutput []string) ([]models.Asset, error)
}
