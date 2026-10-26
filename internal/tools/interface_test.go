package tools_test

import (
	"testing"

	"github.com/diamond/bhf/internal/tools"
)

func TestCleanDomain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"plain domain", "example.com", "example.com"},
		{"wildcard prefix", "*.example.com", "example.com"},
		{"dot prefix", ".example.com", "example.com"},
		{"https prefix", "https://example.com", "example.com"},
		{"http prefix", "http://example.com", "example.com"},
		{"trailing slash", "https://example.com/", "example.com"},
		{"uppercase domain", "EXAMPLE.COM", "example.com"},
		{"leading and trailing spaces", "  sub.example.com/  ", "sub.example.com"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tools.CleanDomain(tc.input)
			if result != tc.expected {
				t.Errorf("CleanDomain(%q) = %q; expected %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestIsInScope(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		target   string
		expected bool
	}{
		{"exact root domain match", "example.com", "example.com", true},
		{"subdomain match", "sub.example.com", "example.com", true},
		{"nested subdomain match", "deep.sub.example.com", "example.com", true},
		{"host with port match", "api.example.com:8443", "example.com", true},
		{"wildcard target definition", "test.example.com", "*.example.com", true},
		{"target with protocol", "test.example.com", "https://example.com", true},
		{"foreign domain out of scope", "attacker.com", "example.com", false},
		{"partial prefix collision out of scope", "notexample.com", "example.com", false},
		{"subdomain with different root", "example.com.evil.com", "example.com", false},
		{"empty host", "", "example.com", false},
		{"empty target", "example.com", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tools.IsInScope(tc.host, tc.target)
			if result != tc.expected {
				t.Errorf("IsInScope(%q, %q) = %v; expected %v", tc.host, tc.target, result, tc.expected)
			}
		})
	}
}
