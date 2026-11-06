package assessment

import (
	"context"
	"net/http"

	"github.com/diamond/bhf/internal/models"
)

// Confidence level constants adhering to the verification philosophy.
const (
	ConfidenceConfirmed     = "CONFIRMED"     // Deterministic proof of exploitability
	ConfidenceHigh          = "HIGH"          // Strong evidence with minimal external variance
	ConfidenceMedium        = "MEDIUM"        // Suspicious behavior requiring manual operator review
	ConfidenceInformational = "INFORMATIONAL" // Defensive posture or configuration observation
)

// AssessmentFinding represents a validated finding with reproducible evidence.
type AssessmentFinding struct {
	RuleID           string `json:"rule_id"`
	Severity         string `json:"severity"`   // critical, high, medium, low, info
	Confidence       string `json:"confidence"` // CONFIRMED, HIGH, MEDIUM, INFORMATIONAL
	Name             string `json:"name"`
	Description      string `json:"description"`
	MatchedAt        string `json:"matched_at"`
	Impact           string `json:"impact"`
	EvidenceRequest  string `json:"evidence_request"`
	EvidenceResponse string `json:"evidence_response"`
	CurlCommand      string `json:"curl_command"`
	Remediation      string `json:"remediation"`
}

// ToFindingModel converts an AssessmentFinding to a persistent database finding entity.
func (af *AssessmentFinding) ToFindingModel(targetID, scanID string) *models.Finding {
	return &models.Finding{
		TargetID:    targetID,
		ScanID:      scanID,
		ToolName:    "bhf-assessment:" + af.RuleID,
		Severity:    models.Severity(af.Severity),
		Name:        af.Name,
		MatchedAt:   af.MatchedAt,
		Description: af.Description + "\n\n**Confidence**: " + af.Confidence + "\n**Impact**: " + af.Impact + "\n**Remediation**: " + af.Remediation,
		CurlCommand: af.CurlCommand,
	}
}

// Rule defines the interface for a deterministic security check.
type Rule interface {
	ID() string
	Name() string
	Category() string
	Description() string
	Evaluate(ctx context.Context, target *models.Target, traffic []models.TrafficRecord, client *http.Client) ([]*AssessmentFinding, error)
}
