package assessment

import (
	"context"
	"net/http"
	"strings"

	"github.com/diamond/bhf/internal/models"
)

// UploadSecurityRule evaluates file upload endpoints for unsafe file validation and execution controls.
type UploadSecurityRule struct{}

func NewUploadSecurityRule() *UploadSecurityRule {
	return &UploadSecurityRule{}
}

func (r *UploadSecurityRule) ID() string {
	return "upload_security"
}

func (r *UploadSecurityRule) Name() string {
	return "File Upload Validation & Access Controls"
}

func (r *UploadSecurityRule) Category() string {
	return "upload"
}

func (r *UploadSecurityRule) Description() string {
	return "Assesses file upload endpoints for dangerous file extension acceptance, MIME-type validation, and script execution exposure."
}

func (r *UploadSecurityRule) Evaluate(ctx context.Context, target *models.Target, traffic []models.TrafficRecord, client *http.Client) ([]*AssessmentFinding, error) {
	findings := make([]*AssessmentFinding, 0)

	for _, rec := range traffic {
		if rec.Category != "upload" && !strings.Contains(rec.ContentType, "multipart/form-data") {
			continue
		}

		lowerReqBody := strings.ToLower(rec.RequestBody)

		// Check for dangerous file extension accepted or observed in traffic
		dangerousExts := []string{".php", ".phtml", ".jsp", ".asp", ".aspx", ".shtml", ".svg", ".html"}
		for _, ext := range dangerousExts {
			if strings.Contains(lowerReqBody, "filename=") && strings.Contains(lowerReqBody, ext) && rec.ResponseStatusCode >= 200 && rec.ResponseStatusCode < 300 {
				req, _ := http.NewRequest(rec.Method, rec.URL, strings.NewReader(rec.RequestBody))
				curl := BuildCurl(req, rec.RequestBody)

				findings = append(findings, &AssessmentFinding{
					RuleID:           r.ID(),
					Severity:         "high",
					Confidence:       ConfidenceConfirmed,
					Name:             "Dangerous File Extension Accepted on Upload Endpoint (" + ext + ")",
					MatchedAt:        rec.URL,
					Description:      "The upload endpoint accepted a file with potentially executable or active content extension (" + ext + ") and returned HTTP " + http.StatusText(rec.ResponseStatusCode) + ".",
					Impact:           "Uploading active script files or SVGs can lead to stored Cross-Site Scripting (XSS), Server-Side Request Forgery (SSRF), or Remote Code Execution (RCE) if served directly.",
					EvidenceRequest:  rec.RequestHeaders + "\n\n" + rec.RequestBody,
					EvidenceResponse: rec.ResponseHeaders + "\n\n" + rec.ResponseBody,
					CurlCommand:      curl,
					Remediation:      "Validate file extensions against a strict whitelist of safe formats (e.g. .png, .jpg, .pdf). Store uploaded files on an isolated CDN or storage bucket without script execution permissions.",
				})
				break
			}
		}
	}

	return findings, nil
}
