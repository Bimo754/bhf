package assessment

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/diamond/bhf/internal/models"
)

// InputValidationRule checks for parameter reflection and special character sanitization.
type InputValidationRule struct{}

func NewInputValidationRule() *InputValidationRule {
	return &InputValidationRule{}
}

func (r *InputValidationRule) ID() string {
	return "input_reflection"
}

func (r *InputValidationRule) Name() string {
	return "Input Reflection & Context Validation"
}

func (r *InputValidationRule) Category() string {
	return "input"
}

func (r *InputValidationRule) Description() string {
	return "Verifies whether user-controlled input parameters reflect unescaped HTML special characters inside response bodies."
}

func (r *InputValidationRule) Evaluate(ctx context.Context, target *models.Target, traffic []models.TrafficRecord, client *http.Client) ([]*AssessmentFinding, error) {
	findings := make([]*AssessmentFinding, 0)
	testedParams := make(map[string]bool)

	for _, rec := range traffic {
		if rec.Method != "GET" || rec.ResponseStatusCode != 200 || client == nil {
			continue
		}

		parsedURL, err := url.Parse(rec.URL)
		if err != nil || parsedURL.RawQuery == "" {
			continue
		}

		query := parsedURL.Query()
		for param := range query {
			paramKey := parsedURL.Path + "?" + param
			if testedParams[paramKey] {
				continue
			}
			testedParams[paramKey] = true

			// Safe canary containing benign test tokens and characters: bhf<test>"'
			canary := "bhf101<canary>test"
			testQuery := parsedURL.Query()
			testQuery.Set(param, canary)
			testURL := *parsedURL
			testURL.RawQuery = testQuery.Encode()

			testReq, _ := http.NewRequestWithContext(ctx, "GET", testURL.String(), nil)
			resp, err := client.Do(testReq)
			if err != nil || resp == nil {
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
				bodyStr := string(bodyBytes)

				// Deterministic Verification: Check if unescaped `<canary>` exists in response body
				if strings.Contains(bodyStr, "<canary>") {
					curl := BuildCurl(testReq, "")
					findings = append(findings, &AssessmentFinding{
						RuleID:           r.ID(),
						Severity:         "medium",
						Confidence:       ConfidenceConfirmed,
						Name:             fmt.Sprintf("Unescaped Parameter Reflection on %q", param),
						MatchedAt:        testURL.String(),
						Description:      fmt.Sprintf("Parameter %q reflects unencoded HTML tags (`<canary>`) directly into the response body. This indicates missing context-aware output encoding.", param),
						Impact:           "Unescaped reflection can lead to Cross-Site Scripting (XSS), enabling session hijacking, DOM manipulation, or credential theft.",
						EvidenceRequest:  FormatRawRequest(testReq, ""),
						EvidenceResponse: FormatRawResponse(resp, bodyStr),
						CurlCommand:      curl,
						Remediation:      "Apply context-aware output encoding (e.g. HTML entity encoding) before reflecting user input in web documents.",
					})
				}
			}
		}
	}

	return findings, nil
}
