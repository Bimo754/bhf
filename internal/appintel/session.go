package appintel

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/diamond/bhf/internal/models"
)

var (
	jwtPattern        = regexp.MustCompile(`eyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,}`)
	bearerPattern     = regexp.MustCompile(`(?i)(Bearer\s+)[A-Za-z0-9_\-\.]{15,}`)
	jsonPassPattern   = regexp.MustCompile(`(?i)("(?:password|passwd|pwd|secret|api_key|token|auth_token)"\s*:\s*)"([^"]+)"`)
	formPassPattern   = regexp.MustCompile(`(?i)((?:password|passwd|pwd|secret|token)=)([^&]+)`)
	creditCardPattern = regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`)
)

// RedactText scrubs passwords, tokens, API keys, and sensitive financial identifiers from strings.
func RedactText(input string) string {
	if input == "" {
		return ""
	}

	out := jwtPattern.ReplaceAllString(input, "[REDACTED_JWT]")
	out = bearerPattern.ReplaceAllString(out, "${1}[REDACTED_TOKEN]")
	out = jsonPassPattern.ReplaceAllString(out, `${1}"[REDACTED]"`)
	out = formPassPattern.ReplaceAllString(out, `${1}[REDACTED]`)
	out = creditCardPattern.ReplaceAllString(out, "[REDACTED_CARD]")

	return out
}

// RedactHeaders produces a sanitized copy of HTTP headers with auth credentials masked.
func RedactHeaders(headers map[string][]string) map[string][]string {
	if headers == nil {
		return make(map[string][]string)
	}

	sanitized := make(map[string][]string, len(headers))
	sensitiveHeaders := map[string]bool{
		"authorization":       true,
		"proxy-authorization": true,
		"cookie":              true,
		"set-cookie":          true,
		"x-api-key":           true,
		"x-auth-token":        true,
		"apikey":              true,
	}

	for k, vals := range headers {
		lowerKey := strings.ToLower(k)
		if sensitiveHeaders[lowerKey] {
			sanitized[k] = []string{"[REDACTED]"}
		} else {
			cleanVals := make([]string, len(vals))
			for i, v := range vals {
				cleanVals[i] = RedactText(v)
			}
			sanitized[k] = cleanVals
		}
	}

	return sanitized
}

// MaskRecord returns a sanitized copy of a TrafficRecord safe for export, logs, or UI viewing.
func MaskRecord(rec *models.TrafficRecord) *models.TrafficRecord {
	if rec == nil {
		return nil
	}

	copyRec := *rec

	// Sanitize request headers
	if rec.RequestHeaders != "" {
		var reqHeaders map[string][]string
		if err := json.Unmarshal([]byte(rec.RequestHeaders), &reqHeaders); err == nil {
			masked := RedactHeaders(reqHeaders)
			if b, err := json.Marshal(masked); err == nil {
				copyRec.RequestHeaders = string(b)
			}
		}
	}

	// Sanitize response headers
	if rec.ResponseHeaders != "" {
		var resHeaders map[string][]string
		if err := json.Unmarshal([]byte(rec.ResponseHeaders), &resHeaders); err == nil {
			masked := RedactHeaders(resHeaders)
			if b, err := json.Marshal(masked); err == nil {
				copyRec.ResponseHeaders = string(b)
			}
		}
	}

	// Sanitize bodies
	copyRec.RequestBody = RedactText(rec.RequestBody)
	copyRec.ResponseBody = RedactText(rec.ResponseBody)

	return &copyRec
}
