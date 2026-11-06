package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/diamond/bhf/internal/appintel"
	"github.com/diamond/bhf/internal/models"
)

// AssessmentProposal represents a test proposed by an AI agent based on traffic analysis.
type AssessmentProposal struct {
	Endpoint  string `json:"endpoint"`
	Method    string `json:"method"`
	TestType  string `json:"test_type"` // auth, idor, input, upload
	Reasoning string `json:"reasoning"`
	Priority  int    `json:"priority"` // 1 (highest) to 5
}

// BuildTrafficAnalysisPrompt compiles a structured summary of intercepted endpoints and parameters for LLM evaluation.
func BuildTrafficAnalysisPrompt(records []models.TrafficRecord, graph *appintel.ParameterGraph) (systemPrompt, userPrompt string) {
	systemPrompt = `You are the BHF Security Assessment Planner. Your task is to analyze intercepted HTTP traffic and parameter graphs from authorized security assessments.
Recommend high-yield, verified security checks (auth, idor, input, upload).
Never hallucinate endpoints or parameters. Ground every proposal strictly in the provided traffic.
Respond ONLY with a JSON object: {"proposals": [{"endpoint": "/path", "method": "GET", "test_type": "idor", "reasoning": "...", "priority": 1}]}`

	var sb strings.Builder
	sb.WriteString("=== INTERCEPTED ENDPOINTS & TRAFFIC SUMMARY ===\n")

	// Limit to distinct endpoints for prompt efficiency
	seen := make(map[string]bool)
	count := 0
	for _, rec := range records {
		key := rec.Method + " " + rec.Path
		if seen[key] || count >= 50 {
			continue
		}
		seen[key] = true
		count++

		sb.WriteString(fmt.Sprintf("- [%s] %s (Category: %s, Status: %d, Auth: %s)\n",
			rec.Method, rec.Path, rec.Category, rec.ResponseStatusCode, rec.AuthContext))
	}

	if graph != nil && len(graph.Parameters) > 0 {
		sb.WriteString("\n=== USER-CONTROLLED PARAMETER GRAPH ===\n")
		paramCount := 0
		for name, entry := range graph.Parameters {
			if paramCount >= 30 {
				break
			}
			paramCount++

			flags := make([]string, 0)
			if entry.IsIDORCandidate {
				flags = append(flags, "IDOR_CANDIDATE")
			}
			if entry.IsAuthParam {
				flags = append(flags, "AUTH_PARAM")
			}

			flagStr := ""
			if len(flags) > 0 {
				flagStr = " [" + strings.Join(flags, ", ") + "]"
			}

			sb.WriteString(fmt.Sprintf("- Parameter %q (Type: %s, Locations: %v)%s\n",
				name, entry.InferredType, entry.Locations, flagStr))
		}
	}

	sb.WriteString("\nIdentify the most critical assessment paths and return the JSON proposals array.")
	return systemPrompt, sb.String()
}

// ParseProposals extracts structured test proposals from an LLM response string.
func ParseProposals(rawResponse string) ([]AssessmentProposal, error) {
	// Clean markdown fences if present
	clean := strings.TrimSpace(rawResponse)
	if strings.HasPrefix(clean, "```json") {
		clean = strings.TrimPrefix(clean, "```json")
		clean = strings.TrimSuffix(clean, "```")
	} else if strings.HasPrefix(clean, "```") {
		clean = strings.TrimPrefix(clean, "```")
		clean = strings.TrimSuffix(clean, "```")
	}
	clean = strings.TrimSpace(clean)

	var wrapper struct {
		Proposals []AssessmentProposal `json:"proposals"`
	}

	if err := json.Unmarshal([]byte(clean), &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse AI proposals JSON: %w (raw: %s)", err, clean)
	}

	return wrapper.Proposals, nil
}
