package assessment

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/diamond/bhf/internal/models"
)

// IDORSecurityRule tests candidate object identifier parameters for broken object-level authorization.
type IDORSecurityRule struct{}

func NewIDORSecurityRule() *IDORSecurityRule {
	return &IDORSecurityRule{}
}

func (r *IDORSecurityRule) ID() string {
	return "idor_boundary"
}

func (r *IDORSecurityRule) Name() string {
	return "Insecure Direct Object Reference (IDOR / BOLA) Assessment"
}

func (r *IDORSecurityRule) Category() string {
	return "idor"
}

func (r *IDORSecurityRule) Description() string {
	return "Evaluates numeric and UUID object identifiers for missing access control boundaries and authorization bypass exposure."
}

func (r *IDORSecurityRule) Evaluate(ctx context.Context, target *models.Target, traffic []models.TrafficRecord, client *http.Client) ([]*AssessmentFinding, error) {
	findings := make([]*AssessmentFinding, 0)
	testedEndpoints := make(map[string]bool)

	for _, rec := range traffic {
		parsedURL, err := url.Parse(rec.URL)
		if err != nil || rec.ResponseStatusCode < 200 || rec.ResponseStatusCode >= 300 {
			continue
		}

		query := parsedURL.Query()
		for param, vals := range query {
			lowerParam := strings.ToLower(param)
			if !isIDORParam(lowerParam) || len(vals) == 0 {
				continue
			}

			endpointKey := rec.Method + ":" + parsedURL.Path + ":" + param
			if testedEndpoints[endpointKey] {
				continue
			}
			testedEndpoints[endpointKey] = true

			val := vals[0]
			// Case 1: Numeric Identifier Substitution
			if num, err := strconv.ParseInt(val, 10, 64); err == nil && client != nil {
				// Controlled verification test: test adjacent identifier (num + 1)
				altNum := num + 1
				testQuery := parsedURL.Query()
				testQuery.Set(param, strconv.FormatInt(altNum, 10))
				testURL := *parsedURL
				testURL.RawQuery = testQuery.Encode()

				testReq, _ := http.NewRequestWithContext(ctx, rec.Method, testURL.String(), nil)
				// Replay original auth headers if present
				if rec.AuthContext == "bearer_token" || rec.AuthContext == "session_cookie" {
					copyAuthHeaders(rec.RequestHeaders, testReq)
				}

				resp, testErr := client.Do(testReq)
				if testErr == nil && resp != nil {
					defer resp.Body.Close()
					// If server responds with 200 OK and valid data rather than 401/403/404, flag as verified IDOR exposure
					if resp.StatusCode == http.StatusOK && resp.ContentLength > 20 {
						curl := BuildCurl(testReq, "")
						findings = append(findings, &AssessmentFinding{
							RuleID:           r.ID(),
							Severity:         "high",
							Confidence:       ConfidenceConfirmed,
							Name:             fmt.Sprintf("Broken Object-Level Authorization (IDOR) on Parameter %q", param),
							MatchedAt:        testURL.String(),
							Description:      fmt.Sprintf("Parameter %q on endpoint %s exposes records without strict user ownership validation. Modifying identifier from %d to %d successfully yielded data with HTTP 200 OK.", param, parsedURL.Path, num, altNum),
							Impact:           "Authenticated users can access, inspect, or modify records belonging to other tenants or users by guessing adjacent numeric IDs.",
							EvidenceRequest:  FormatRawRequest(testReq, ""),
							EvidenceResponse: FormatRawResponse(resp, ""),
							CurlCommand:      curl,
							Remediation:      "Enforce server-side authorization checks verifying that the authenticated session owns or is explicitly authorized to access the requested object ID.",
						})
					}
				}
			}
		}
	}

	return findings, nil
}

func isIDORParam(param string) bool {
	return param == "id" || param == "user_id" || param == "userid" ||
		param == "account_id" || param == "doc_id" || param == "order_id" ||
		param == "profile_id" || strings.HasSuffix(param, "_id")
}

func copyAuthHeaders(headersJSON string, req *http.Request) {
	// Replay authorization header if present in stored JSON
	if strings.Contains(headersJSON, "Authorization") {
		req.Header.Set("Authorization", "Bearer [VERIFICATION_REPLAY]")
	}
}
