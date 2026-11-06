package assessment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/diamond/bhf/internal/models"
)

// AuthSecurityRule audits authentication workflows for credentials exposed in URLs and insecure cookies.
type AuthSecurityRule struct{}

func NewAuthSecurityRule() *AuthSecurityRule {
	return &AuthSecurityRule{}
}

func (r *AuthSecurityRule) ID() string {
	return "auth_security"
}

func (r *AuthSecurityRule) Name() string {
	return "Authentication & Session Security Assessment"
}

func (r *AuthSecurityRule) Category() string {
	return "auth"
}

func (r *AuthSecurityRule) Description() string {
	return "Audits authentication requests for credential exposure in URLs, insecure session cookie attributes, and missing transport protections."
}

func (r *AuthSecurityRule) Evaluate(ctx context.Context, target *models.Target, traffic []models.TrafficRecord, client *http.Client) ([]*AssessmentFinding, error) {
	findings := make([]*AssessmentFinding, 0)

	for _, rec := range traffic {
		parsedURL, err := url.Parse(rec.URL)
		if err != nil {
			continue
		}

		// Check 1: Credentials or Passwords in Query String (GET requests)
		if rec.Method == "GET" && parsedURL.RawQuery != "" {
			query := parsedURL.Query()
			for param := range query {
				lower := strings.ToLower(param)
				if lower == "password" || lower == "passwd" || lower == "pwd" || lower == "secret" || lower == "access_token" || lower == "api_key" || lower == "apikey" || lower == "token" {
					req, _ := http.NewRequest(rec.Method, rec.URL, nil)
					curl := BuildCurl(req, "")

					findings = append(findings, &AssessmentFinding{
						RuleID:           r.ID(),
						Severity:         "high",
						Confidence:       ConfidenceConfirmed,
						Name:             "Sensitive Credentials Exposed in URL Query Parameters",
						MatchedAt:        rec.URL,
						Description:      "The application accepts sensitive credentials or access tokens within URL query parameters. URLs are logged in web server access logs, browser history, proxy logs, and Referer headers.",
						Impact:           "Adversaries with access to server logs, browser caches, or upstream proxies can intercept plaintext user credentials.",
						EvidenceRequest:  rec.RequestHeaders + "\n\nURL: " + rec.URL,
						EvidenceResponse: rec.ResponseHeaders,
						CurlCommand:      curl,
						Remediation:      "Transmit credentials exclusively within request bodies using POST/PUT requests over TLS with application/json or form encoding.",
					})
				}
			}
		}

		// Check 2: Session Cookies Missing 'Secure' or 'HttpOnly' Flags
		if rec.ResponseHeaders != "" {
			var headers map[string][]string
			if err := json.Unmarshal([]byte(rec.ResponseHeaders), &headers); err == nil {
				cookies := headers["Set-Cookie"]
				for _, cookie := range cookies {
					lowerCookie := strings.ToLower(cookie)
					isSessionCookie := strings.Contains(lowerCookie, "sess") || strings.Contains(lowerCookie, "token") ||
						strings.Contains(lowerCookie, "auth") || strings.Contains(lowerCookie, "jwt")

					if isSessionCookie {
						missingFlags := make([]string, 0)
						if !strings.Contains(lowerCookie, "httponly") {
							missingFlags = append(missingFlags, "HttpOnly")
						}
						if !strings.Contains(lowerCookie, "secure") && strings.HasPrefix(rec.URL, "https://") {
							missingFlags = append(missingFlags, "Secure")
						}

						if len(missingFlags) > 0 {
							req, _ := http.NewRequest(rec.Method, rec.URL, nil)
							curl := BuildCurl(req, "")

							findings = append(findings, &AssessmentFinding{
								RuleID:           r.ID(),
								Severity:         "medium",
								Confidence:       ConfidenceConfirmed,
								Name:             "Session Cookie Missing Security Flags (" + strings.Join(missingFlags, ", ") + ")",
								MatchedAt:        rec.URL,
								Description:      "The application issues session cookies without essential defensive flags: " + strings.Join(missingFlags, ", ") + ".\nCookie: `" + cookie + "`",
								Impact:           "Missing HttpOnly permits client-side scripts to steal session tokens via XSS. Missing Secure permits transmission over unencrypted channels.",
								EvidenceRequest:  rec.RequestHeaders,
								EvidenceResponse: "Set-Cookie: " + cookie,
								CurlCommand:      curl,
								Remediation:      "Ensure all session management cookies include 'HttpOnly', 'Secure', and 'SameSite=Lax' (or 'Strict') attributes.",
							})
						}
					}
				}
			}
		}
	}

	return findings, nil
}
