package network

import (
	"context"
	"fmt"
	"strings"

	"github.com/diamond/bhf/internal/models"
)

// NetworkMisconfigResult represents identified infrastructure exposures.
type NetworkMisconfigResult struct {
	ServiceName string `json:"service_name"`
	Port        int    `json:"port"`
	Severity    string `json:"severity"` // critical, high, medium, low
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

// AuditNetworkService evaluates open network services for severe exposure misconfigurations.
func AuditNetworkService(ctx context.Context, svc models.NetworkService) []NetworkMisconfigResult {
	results := make([]NetworkMisconfigResult, 0)

	switch svc.Port {
	case 23:
		results = append(results, NetworkMisconfigResult{
			ServiceName: svc.Name,
			Port:        svc.Port,
			Severity:    "high",
			Title:       "Insecure Cleartext Telnet Management Protocol Active",
			Description: "Telnet operates on port 23 without cryptographic transport encryption. All authentication credentials and session keystrokes are transmitted in plaintext.",
			Impact:      "Adversaries on the network path can sniff administrative passwords and inject commands.",
		})

	case 21:
		results = append(results, NetworkMisconfigResult{
			ServiceName: svc.Name,
			Port:        svc.Port,
			Severity:    "medium",
			Title:       "Cleartext FTP File Transfer Protocol Exposed",
			Description: "FTP service on port 21 transmits user credentials and file payloads in plaintext without TLS/SSL encryption.",
			Impact:      "Credentials and sensitive transferred documents are subject to network eavesdropping.",
		})

	case 6379:
		if strings.HasPrefix(svc.Banner, "+PONG") {
			results = append(results, NetworkMisconfigResult{
				ServiceName: "redis",
				Port:        svc.Port,
				Severity:    "critical",
				Title:       "Unauthenticated Redis Key-Value Store Exposed",
				Description: fmt.Sprintf("The Redis instance on port 6379 responded with %q without requiring password authentication.", svc.Banner),
				Impact:      "Unauthenticated attackers can read, modify, or flush all cached memory data, and potentially execute system commands or write authorized SSH keys.",
			})
		}

	case 9200:
		results = append(results, NetworkMisconfigResult{
			ServiceName: "elasticsearch",
			Port:        svc.Port,
			Severity:    "high",
			Title:       "Elasticsearch Cluster API Service Exposed",
			Description: "Elasticsearch REST API on port 9200 is reachable across the network perimeter.",
			Impact:      "Unauthenticated access may allow extraction of indexed database records, cluster modification, or data exfiltration.",
		})

	case 27017:
		results = append(results, NetworkMisconfigResult{
			ServiceName: "mongodb",
			Port:        svc.Port,
			Severity:    "high",
			Title:       "MongoDB Database Service Exposed",
			Description: "MongoDB database listener on port 27017 is directly accessible on the network interface.",
			Impact:      "Exposing database listeners externally increases exposure to brute force attacks, authentication bypasses, or ransomware.",
		})
	}

	return results
}
