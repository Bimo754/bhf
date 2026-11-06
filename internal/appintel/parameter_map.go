package appintel

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/diamond/bhf/internal/models"
)

var (
	uuidRegex  = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// ParameterEntry represents an aggregated user-controlled parameter discovered across traffic.
type ParameterEntry struct {
	Name            string   `json:"name"`
	Locations       []string `json:"locations"` // query, body, json, header
	Endpoints       []string `json:"endpoints"` // Distinct paths where parameter appears
	SampleValues    []string `json:"sample_values"`
	InferredType    string   `json:"inferred_type"` // integer, uuid, email, boolean, token, string
	IsIDORCandidate bool     `json:"is_idor_candidate"`
	IsAuthParam     bool     `json:"is_auth_param"`
	OccurrenceCount int      `json:"occurrence_count"`
}

// ParameterGraph maps parameter names to their aggregated entry analysis.
type ParameterGraph struct {
	Parameters map[string]*ParameterEntry `json:"parameters"`
	TotalCount int                        `json:"total_count"`
}

// BuildParameterGraph aggregates and classifies parameters from a slice of traffic records.
func BuildParameterGraph(records []models.TrafficRecord) *ParameterGraph {
	graph := &ParameterGraph{
		Parameters: make(map[string]*ParameterEntry),
	}

	for _, rec := range records {
		parsedURL, _ := url.Parse(rec.URL)
		path := rec.Path
		if path == "" && parsedURL != nil {
			path = parsedURL.Path
		}

		// 1. Process Query Parameters
		if parsedURL != nil {
			for key, vals := range parsedURL.Query() {
				val := ""
				if len(vals) > 0 {
					val = vals[0]
				}
				addParamToGraph(graph, key, "query", path, val)
			}
		}

		// 2. Process Body Parameters
		if rec.RequestBody != "" {
			if strings.Contains(rec.ContentType, "application/json") {
				var jsonMap map[string]any
				if err := json.Unmarshal([]byte(rec.RequestBody), &jsonMap); err == nil {
					flattenJSON(jsonMap, "", func(key, val string) {
						addParamToGraph(graph, key, "json", path, val)
					})
				}
			} else if strings.Contains(rec.ContentType, "application/x-www-form-urlencoded") {
				if vals, err := url.ParseQuery(rec.RequestBody); err == nil {
					for key, list := range vals {
						val := ""
						if len(list) > 0 {
							val = list[0]
						}
						addParamToGraph(graph, key, "body", path, val)
					}
				}
			}
		}
	}

	graph.TotalCount = len(graph.Parameters)
	return graph
}

func addParamToGraph(graph *ParameterGraph, name, location, path, sampleVal string) {
	entry, exists := graph.Parameters[name]
	if !exists {
		inferred := inferType(sampleVal)
		entry = &ParameterEntry{
			Name:            name,
			Locations:       []string{location},
			Endpoints:       []string{path},
			SampleValues:    make([]string, 0),
			InferredType:    inferred,
			IsIDORCandidate: isIDORCandidate(name, inferred),
			IsAuthParam:     isAuthParam(name),
			OccurrenceCount: 1,
		}
		if sampleVal != "" {
			entry.SampleValues = append(entry.SampleValues, sampleVal)
		}
		graph.Parameters[name] = entry
		return
	}

	entry.OccurrenceCount++

	if !containsString(entry.Locations, location) {
		entry.Locations = append(entry.Locations, location)
	}
	if path != "" && !containsString(entry.Endpoints, path) {
		entry.Endpoints = append(entry.Endpoints, path)
	}
	if sampleVal != "" && len(entry.SampleValues) < 5 && !containsString(entry.SampleValues, sampleVal) {
		entry.SampleValues = append(entry.SampleValues, sampleVal)
	}

	// Refine inferred type if previously set to string but better candidate found
	if entry.InferredType == "string" && sampleVal != "" {
		entry.InferredType = inferType(sampleVal)
		entry.IsIDORCandidate = isIDORCandidate(name, entry.InferredType)
	}
}

func inferType(val string) string {
	if val == "" {
		return "string"
	}
	if val == "true" || val == "false" || val == "1" || val == "0" {
		return "boolean"
	}
	if _, err := strconv.ParseInt(val, 10, 64); err == nil {
		return "integer"
	}
	if uuidRegex.MatchString(val) {
		return "uuid"
	}
	if emailRegex.MatchString(val) {
		return "email"
	}
	if len(val) >= 32 && (strings.HasPrefix(val, "eyJ") || isHexOrBase64(val)) {
		return "token"
	}
	return "string"
}

func isIDORCandidate(name, inferredType string) bool {
	lower := strings.ToLower(name)
	idorKeywords := []string{
		"id", "user_id", "userid", "account_id", "accountid",
		"org_id", "organization_id", "customer_id", "doc_id",
		"document_id", "profile_id", "uuid", "order_id", "invoice_id",
	}

	for _, kw := range idorKeywords {
		if lower == kw || strings.HasSuffix(lower, "_id") || strings.HasSuffix(lower, "id") {
			return true
		}
	}

	// Non-trivial integer or UUID parameters are high probability IDOR candidates
	if (inferredType == "integer" || inferredType == "uuid") && (strings.Contains(lower, "num") || strings.Contains(lower, "ref") || strings.Contains(lower, "key")) {
		return true
	}

	return false
}

func isAuthParam(name string) bool {
	lower := strings.ToLower(name)
	authKeywords := []string{
		"password", "passwd", "pwd", "secret", "token", "jwt",
		"api_key", "apikey", "session", "csrf", "auth", "code", "pin",
	}
	for _, kw := range authKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func flattenJSON(data map[string]any, prefix string, callback func(key, val string)) {
	for k, v := range data {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}

		switch val := v.(type) {
		case string:
			callback(fullKey, val)
		case float64:
			callback(fullKey, strconv.FormatFloat(val, 'f', -1, 64))
		case bool:
			callback(fullKey, strconv.FormatBool(val))
		case map[string]any:
			flattenJSON(val, fullKey, callback)
		}
	}
}

func containsString(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}

func isHexOrBase64(s string) bool {
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' || c == '=' || c == '+' || c == '/') {
			return false
		}
	}
	return true
}
