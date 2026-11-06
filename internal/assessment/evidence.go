package assessment

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// FormatRawRequest serializes an HTTP request into standard RFC 7230 wire format.
func FormatRawRequest(req *http.Request, body string) string {
	if req == nil {
		return ""
	}

	var sb strings.Builder
	path := req.URL.RequestURI()
	if path == "" {
		path = "/"
	}

	sb.WriteString(fmt.Sprintf("%s %s HTTP/1.1\r\n", req.Method, path))
	sb.WriteString(fmt.Sprintf("Host: %s\r\n", req.Host))

	// Sort headers for deterministic output
	headerKeys := make([]string, 0, len(req.Header))
	for k := range req.Header {
		headerKeys = append(headerKeys, k)
	}
	sort.Strings(headerKeys)

	for _, k := range headerKeys {
		for _, v := range req.Header[k] {
			sb.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}

	sb.WriteString("\r\n")
	if body != "" {
		sb.WriteString(body)
	}

	return sb.String()
}

// FormatRawResponse serializes an HTTP response into standard RFC 7230 wire format.
func FormatRawResponse(resp *http.Response, body string) string {
	if resp == nil {
		return ""
	}

	var sb strings.Builder
	proto := resp.Proto
	if proto == "" {
		proto = "HTTP/1.1"
	}

	sb.WriteString(fmt.Sprintf("%s %s\r\n", proto, resp.Status))

	headerKeys := make([]string, 0, len(resp.Header))
	for k := range resp.Header {
		headerKeys = append(headerKeys, k)
	}
	sort.Strings(headerKeys)

	for _, k := range headerKeys {
		for _, v := range resp.Header[k] {
			sb.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}

	sb.WriteString("\r\n")
	if body != "" {
		// Limit response body in evidence if excessive
		if len(body) > 2048 {
			sb.WriteString(body[:2048] + "\n... [TRUNCATED]")
		} else {
			sb.WriteString(body)
		}
	}

	return sb.String()
}

// BuildCurl generates an exact, copy-pasteable curl reproduction command.
func BuildCurl(req *http.Request, body string) string {
	if req == nil || req.URL == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("curl -i -s -k")

	if req.Method != "" && req.Method != "GET" {
		sb.WriteString(fmt.Sprintf(" -X %s", req.Method))
	}

	// Sort headers
	headerKeys := make([]string, 0, len(req.Header))
	for k := range req.Header {
		headerKeys = append(headerKeys, k)
	}
	sort.Strings(headerKeys)

	for _, k := range headerKeys {
		for _, v := range req.Header[k] {
			sb.WriteString(fmt.Sprintf(" -H %q", fmt.Sprintf("%s: %s", k, v)))
		}
	}

	if body != "" {
		escaped := strings.ReplaceAll(body, "'", `'\''`)
		sb.WriteString(fmt.Sprintf(" --data-raw '%s'", escaped))
	}

	sb.WriteString(fmt.Sprintf(" %q", req.URL.String()))
	return sb.String()
}

// CompileEvidence constructs the complete evidence package for a security finding.
func CompileEvidence(req *http.Request, reqBody string, resp *http.Response, respBody string) (rawReq, rawResp, curl string) {
	rawReq = FormatRawRequest(req, reqBody)
	rawResp = FormatRawResponse(resp, respBody)
	curl = BuildCurl(req, reqBody)
	return rawReq, rawResp, curl
}
