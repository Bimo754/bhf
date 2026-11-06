package appintel_test

import (
	"strings"
	"testing"

	"github.com/diamond/bhf/internal/appintel"
	"github.com/diamond/bhf/internal/models"
)

func TestCategorize(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		url              string
		body             string
		contentType      string
		expectedCategory string
	}{
		{
			name:             "login endpoint",
			method:           "POST",
			url:              "https://example.com/api/v1/auth/login",
			body:             `{"username":"admin"}`,
			contentType:      "application/json",
			expectedCategory: appintel.CategoryAuth,
		},
		{
			name:             "registration endpoint",
			method:           "POST",
			url:              "https://example.com/signup",
			body:             `{"email":"a@b.com"}`,
			contentType:      "application/json",
			expectedCategory: appintel.CategoryRegistration,
		},
		{
			name:             "password reset endpoint",
			method:           "POST",
			url:              "https://example.com/forgot-password",
			body:             `{"email":"a@b.com"}`,
			contentType:      "application/x-www-form-urlencoded",
			expectedCategory: appintel.CategoryAccountRecovery,
		},
		{
			name:             "file upload multipart",
			method:           "POST",
			url:              "https://example.com/profile/avatar",
			body:             "BINARY_DATA",
			contentType:      "multipart/form-data; boundary=---123",
			expectedCategory: appintel.CategoryUpload,
		},
		{
			name:             "search query endpoint",
			method:           "GET",
			url:              "https://example.com/search?q=test&page=1",
			body:             "",
			contentType:      "text/html",
			expectedCategory: appintel.CategorySearch,
		},
		{
			name:             "user generated comments",
			method:           "POST",
			url:              "https://example.com/articles/123/comments",
			body:             `{"comment":"hello"}`,
			contentType:      "application/json",
			expectedCategory: appintel.CategoryUGC,
		},
		{
			name:             "administrative metrics",
			method:           "GET",
			url:              "https://example.com/admin/metrics",
			body:             "",
			contentType:      "application/json",
			expectedCategory: appintel.CategoryAdmin,
		},
		{
			name:             "api data read",
			method:           "GET",
			url:              "https://example.com/api/v2/products",
			body:             "",
			contentType:      "application/json",
			expectedCategory: appintel.CategoryAPI,
		},
		{
			name:             "general static page",
			method:           "GET",
			url:              "https://example.com/about-us",
			body:             "",
			contentType:      "text/html",
			expectedCategory: appintel.CategoryGeneral,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := appintel.Categorize(tc.method, tc.url, nil, tc.body, tc.contentType)
			if res.Category != tc.expectedCategory {
				t.Errorf("Categorize(%q) = %q; want %q (reasons: %v)", tc.url, res.Category, tc.expectedCategory, res.Reasons)
			}
		})
	}
}

func TestParameterGraph(t *testing.T) {
	records := []models.TrafficRecord{
		{
			URL:         "https://example.com/api/users?user_id=105&format=json",
			Method:      "GET",
			Path:        "/api/users",
			ContentType: "application/json",
		},
		{
			URL:         "https://example.com/api/users",
			Method:      "POST",
			Path:        "/api/users",
			ContentType: "application/json",
			RequestBody: `{"account_id": "c7325a7b-3b6b-4e89-9a25-c6e7f1e62689", "email": "alice@example.com", "password": "SuperSecretPassword123!"}`,
		},
	}

	graph := appintel.BuildParameterGraph(records)
	if graph.TotalCount < 4 {
		t.Fatalf("expected at least 4 parameters in graph, got %d", graph.TotalCount)
	}

	// Verify user_id is an IDOR candidate with integer type
	userIDEntry := graph.Parameters["user_id"]
	if userIDEntry == nil {
		t.Fatal("expected user_id in parameter graph")
	}
	if !userIDEntry.IsIDORCandidate {
		t.Errorf("expected user_id to be flagged as IDOR candidate")
	}
	if userIDEntry.InferredType != "integer" {
		t.Errorf("expected user_id inferred type integer, got %s", userIDEntry.InferredType)
	}

	// Verify account_id is an IDOR candidate with uuid type
	accountIDEntry := graph.Parameters["account_id"]
	if accountIDEntry == nil {
		t.Fatal("expected account_id in parameter graph")
	}
	if !accountIDEntry.IsIDORCandidate {
		t.Errorf("expected account_id to be flagged as IDOR candidate")
	}
	if accountIDEntry.InferredType != "uuid" {
		t.Errorf("expected account_id inferred type uuid, got %s", accountIDEntry.InferredType)
	}

	// Verify password is flagged as auth parameter
	passEntry := graph.Parameters["password"]
	if passEntry == nil {
		t.Fatal("expected password in parameter graph")
	}
	if !passEntry.IsAuthParam {
		t.Errorf("expected password to be flagged as auth parameter")
	}
}

func TestSensitiveDataRedaction(t *testing.T) {
	rawJSON := `{"username":"bob","password":"TopSecretPassword99","jwt":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozGz4tCPsmunit"}`
	clean := appintel.RedactText(rawJSON)

	if strings.Contains(clean, "TopSecretPassword99") {
		t.Errorf("password was not redacted: %s", clean)
	}
	if strings.Contains(clean, "eyJhbGciOiJIUzI1Ni") {
		t.Errorf("JWT token was not redacted: %s", clean)
	}
	if !strings.Contains(clean, "[REDACTED]") {
		t.Errorf("expected [REDACTED] token in output: %s", clean)
	}

	// Test Header Redaction
	headers := map[string][]string{
		"Authorization": {"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.abcdefg1234567890"},
		"Cookie":        {"sessionid=secret_session_token_123"},
		"User-Agent":    {"BHF/1.0"},
	}

	masked := appintel.RedactHeaders(headers)
	if masked["Authorization"][0] != "[REDACTED]" {
		t.Errorf("Authorization header not masked: %v", masked["Authorization"])
	}
	if masked["Cookie"][0] != "[REDACTED]" {
		t.Errorf("Cookie header not masked: %v", masked["Cookie"])
	}
	if masked["User-Agent"][0] != "BHF/1.0" {
		t.Errorf("User-Agent should remain intact: %v", masked["User-Agent"])
	}
}
