package appintel

import (
	"encoding/json"
	"net/url"
	"strings"
)

// Standard functional categories for application traffic.
const (
	CategoryAuth            = "auth"
	CategoryRegistration    = "registration"
	CategoryAccountRecovery = "account_recovery"
	CategoryUpload          = "upload"
	CategorySearch          = "search"
	CategoryUGC             = "ugc"
	CategoryAdmin           = "admin"
	CategoryAPI             = "api"
	CategoryGeneral         = "general"
)

// CategorizationResult represents the inferred functionality of an HTTP transaction.
type CategorizationResult struct {
	Category    string   `json:"category"`
	Subcategory string   `json:"subcategory"`
	Confidence  float64  `json:"confidence"` // 0.0 to 1.0
	Reasons     []string `json:"reasons"`
}

// Categorize analyzes an HTTP request and assigns a functional classification with confidence and reasoning.
func Categorize(method, rawURL string, headers map[string][]string, body string, contentType string) CategorizationResult {
	parsedURL, err := url.Parse(rawURL)
	var path string
	if err == nil {
		path = strings.ToLower(parsedURL.Path)
	} else {
		path = strings.ToLower(rawURL)
	}

	lowerContentType := strings.ToLower(contentType)
	lowerBody := strings.ToLower(body)

	// 1. Account Recovery (Password Reset)
	if matchesAny(path, "forgot-password", "reset-password", "password/reset", "recover", "account-recovery") {
		return CategorizationResult{
			Category:    CategoryAccountRecovery,
			Subcategory: "password_reset",
			Confidence:  0.95,
			Reasons:     []string{"path contains password reset keywords"},
		}
	}

	// 2. Registration
	if matchesAny(path, "register", "signup", "sign-up", "create-account", "new-user", "verify-email") {
		return CategorizationResult{
			Category:    CategoryRegistration,
			Subcategory: "account_creation",
			Confidence:  0.95,
			Reasons:     []string{"path contains account registration keywords"},
		}
	}

	// 3. Authentication (Login / Logout / Tokens / MFA)
	if matchesAny(path, "login", "signin", "sign-in", "logout", "signout", "oauth", "token", "saml", "2fa", "mfa", "auth") {
		subcat := "login"
		if strings.Contains(path, "logout") || strings.Contains(path, "signout") {
			subcat = "logout"
		} else if strings.Contains(path, "oauth") || strings.Contains(path, "token") {
			subcat = "token_exchange"
		} else if strings.Contains(path, "2fa") || strings.Contains(path, "mfa") {
			subcat = "multi_factor"
		}
		return CategorizationResult{
			Category:    CategoryAuth,
			Subcategory: subcat,
			Confidence:  0.90,
			Reasons:     []string{"path matches authentication endpoints"},
		}
	}

	// Also check body for login parameters if path was generic (e.g. POST /api/v1/session)
	if method == "POST" && strings.Contains(lowerBody, "password") && (strings.Contains(lowerBody, "username") || strings.Contains(lowerBody, "email")) {
		return CategorizationResult{
			Category:    CategoryAuth,
			Subcategory: "credential_submission",
			Confidence:  0.85,
			Reasons:     []string{"request body contains username and password fields"},
		}
	}

	// 4. File Operations & Uploads
	if strings.Contains(lowerContentType, "multipart/form-data") || matchesAny(path, "upload", "attachment", "avatar", "import", "media", "document") {
		return CategorizationResult{
			Category:    CategoryUpload,
			Subcategory: "file_upload",
			Confidence:  0.90,
			Reasons:     []string{"multipart content-type or upload path detected"},
		}
	}

	// 5. Search & Discovery
	if matchesAny(path, "search", "query", "filter", "find", "suggest", "autocomplete") || (parsedURL != nil && (parsedURL.Query().Has("q") || parsedURL.Query().Has("search") || parsedURL.Query().Has("query"))) {
		return CategorizationResult{
			Category:    CategorySearch,
			Subcategory: "query_filter",
			Confidence:  0.85,
			Reasons:     []string{"search path or search query parameter detected"},
		}
	}

	// 6. User Generated Content (UGC)
	if matchesAny(path, "comment", "review", "message", "feedback", "post", "ticket", "article") {
		subcat := "content_read"
		if method == "POST" || method == "PUT" {
			subcat = "content_creation"
		}
		return CategorizationResult{
			Category:    CategoryUGC,
			Subcategory: subcat,
			Confidence:  0.80,
			Reasons:     []string{"path matches user-generated content resource"},
		}
	}

	// 7. Administrative & Internal Controls
	if matchesAny(path, "admin", "dashboard", "manage", "management", "metrics", "actuator", "internal", "config", "debug") {
		return CategorizationResult{
			Category:    CategoryAdmin,
			Subcategory: "administrative_panel",
			Confidence:  0.90,
			Reasons:     []string{"path contains administrative keywords"},
		}
	}

	// 8. API Operations
	if strings.Contains(path, "/api/") || strings.Contains(lowerContentType, "application/json") || strings.Contains(lowerContentType, "application/xml") || method == "PUT" || method == "PATCH" || method == "DELETE" {
		subcat := "rest_api"
		if strings.Contains(lowerBody, "query") && strings.Contains(lowerBody, "mutation") {
			subcat = "graphql"
		}
		return CategorizationResult{
			Category:    CategoryAPI,
			Subcategory: subcat,
			Confidence:  0.85,
			Reasons:     []string{"API path, structured content-type, or state-modifying verb"},
		}
	}

	// 9. General Fallback
	return CategorizationResult{
		Category:    CategoryGeneral,
		Subcategory: "standard_web",
		Confidence:  0.50,
		Reasons:     []string{"default web resource classification"},
	}
}

// CategorizeJSON parses headers and body from stored JSON strings on TrafficRecord.
func CategorizeJSON(method, rawURL, reqHeadersJSON, body, contentType string) CategorizationResult {
	var headers map[string][]string
	if reqHeadersJSON != "" {
		_ = json.Unmarshal([]byte(reqHeadersJSON), &headers)
	}
	return Categorize(method, rawURL, headers, body, contentType)
}

func matchesAny(text string, targets ...string) bool {
	for _, target := range targets {
		if strings.Contains(text, target) {
			return true
		}
	}
	return false
}
