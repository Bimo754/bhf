# BHF — Bug Hunting & Security Assessment Platform: Testing Strategy and Guidelines

This document outlines the testing architecture, testing principles, suite structure, and implementation standards for BHF. It serves as the definitive reference for maintaining high test coverage and preventing regressions without introducing brittle or bloated test code.

---

## 1. Testing Philosophy

- **Deterministic and Fast**: Tests must execute in milliseconds without external network calls, real port scans, or live web requests against target domains.
- **Isolated State**: Every database test must operate against a dedicated temporary SQLite database (`t.TempDir()`) or an in-memory instance (`:memory:`), ensuring complete test isolation and zero cross-test interference.
- **Evidence Verification Testing**: Tests for assessment checks must evaluate deterministic pass/fail rules against recorded HTTP fixtures (requests and responses), ensuring that findings are generated strictly on verified evidence.
- **Controlled Growth (No Exponential Bloat)**:
  - Test business logic, parsing accuracy, scoping boundaries, state transitions, and verification rules.
  - Avoid writing hundreds of identical redundant tests for simple struct getters.
  - Prefer clean, table-driven tests (`[]struct{ name, input, expected }`) that test multiple scenarios within a single, readable test function.
- **Zero Flakiness**: Tests must not rely on `time.Sleep()` or arbitrary delays. Synchronize goroutines using channels, wait groups (`sync.WaitGroup`), or context cancellations.
- **Zero Emojis**: Never use emojis in test function names, log statements, or assertion failure messages.

---

## 2. Test Architecture by Layer

### 2.1 Database Repositories (`internal/database/`)

Tests in this package verify SQLite migrations, CRUD operations, foreign key cascades, and complex analytical queries (such as asset prioritization, endpoint categorization, and finding statistics).

**Key Rules**:
- Initialize a fresh database with `filepath.Join(t.TempDir(), "test.db")` or `:memory:`.
- Execute `MigrateSchema(db)` before testing repositories.
- Verify cascading deletes: deleting a target must automatically clean up associated scans, assets, endpoints, findings, and logs.

**Reference Implementations**:
- `internal/database/asset_repo_test.go`: Verifies asset upserts, conflict resolution (`UNIQUE(target_id, host, port)`), and priority score calculation.
- `internal/database/endpoint_repo_test.go`: Verifies endpoint insertion, route queries, status code filtering, and deduplication (`UNIQUE(target_id, url)`).
- `internal/database/finding_repo_test.go`: Verifies finding insertions, severity ordering, reproducible evidence persistence, and aggregate count statistics.

### 2.2 Security Tools & Parsers (`internal/tools/`)

External tool wrappers wrap CLI binaries. Tool tests must NEVER execute real binaries or send live packets. Instead, test the parsing engine and scope filtering logic:

1. **Scoping and Domain Sanitization**:
   - `CleanDomain()`: Verify stripping of `*.`, `https://`, `http://`, ports, and trailing slashes.
   - `IsInScope()`: Verify that target subdomains, CIDRs, and IPs are in scope while foreign domains, third-party CDNs, and out-of-scope targets are strictly rejected.
2. **Parser Unit Tests**:
   - Test `Tool.Run(execCtx, rawOutput)` by supplying fixed strings representing CLI output or JSON lines.
   - Test error handling when output files or stdout lines contain malformed JSON or empty strings.

### 2.3 Application Intelligence & Traffic Analysis (`internal/appintel/`)

Tests in this package verify HTTP traffic interception, transaction normalization, and traffic classification:
- **Traffic Categorization**: Verify that login forms, registration endpoints, password resets, file upload endpoints, search forms, and API routes are accurately classified based on path, headers, and request body structures.
- **Parameter Extraction**: Verify that query parameters, JSON payload keys, and form fields are mapped cleanly into parameter graphs without leaking sensitive tokens or PII.

### 2.4 Intelligent Security Assessment Engine (`internal/assessment/`)

Tests in this package verify deterministic security checks and verification rules:
- **Verification Rule Accuracy**: Test security check evaluations against fixed HTTP request/response fixture pairs.
- **False Positive Resistance**: Ensure that mere error responses (e.g. 500 Internal Server Error), generic 403s, or standard reflected query parameters without security impact do not trigger finding records.
- **Evidence Formatting**: Verify that generated findings include complete `evidence_request`, `evidence_response`, and a valid, reproducible `curl_command`.

### 2.5 HTTP Presentation Layer (`internal/api/`)

API handlers are tested using Go's standard `net/http/httptest` package:
- Use `httptest.NewRecorder()` to capture HTTP responses.
- Construct `http.Request` instances via `httptest.NewRequest()`.
- Inject repository instances backed by an in-memory SQLite database.
- Assert HTTP status codes, `Content-Type` headers, and JSON response envelopes.

### 2.6 Model Context Protocol (`internal/mcp/`)

Tests in `internal/mcp/server_test.go` verify JSON-RPC 2.0 protocol conformance:
- `initialize`: Verify protocol version (`2024-11-05`) and server capabilities.
- `tools/list`: Verify that all registered tools and their input schemas are reported.
- `tools/call`: Verify tool execution (e.g., `list_targets`, `list_endpoints`, `get_target_stats`) and structured error responses on invalid parameters.

---

## 3. How to Run Tests

### Running Backend Go Tests

Execute all test packages in the repository:

```bash
go test -v ./...
```

Run tests with data race detection enabled:

```bash
go test -race ./...
```

Run tests with code coverage analysis:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Run tests for a specific package:

```bash
go test -v ./internal/database
go test -v ./internal/tools
```

### Verifying Frontend Build

Verify that the React frontend builds without TypeScript or bundling errors:

```bash
cd web
npm run build
```

---

## 4. Standard Table-Driven Test Patterns

### Pattern A: Scoping and String Sanitization

```go
package tools_test

import (
	"testing"
	"github.com/diamond/bhf/internal/tools"
)

func TestIsInScope(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		target   string
		expected bool
	}{
		{"exact root domain", "example.com", "example.com", true},
		{"subdomain", "api.example.com", "example.com", true},
		{"nested subdomain", "v1.auth.example.com", "example.com", true},
		{"host with port", "api.example.com:8443", "example.com", true},
		{"wildcard root target", "sub.example.com", "*.example.com", true},
		{"out of scope domain", "evil.com", "example.com", false},
		{"partial prefix match", "notexample.com", "example.com", false},
		{"empty hostname", "", "example.com", false},
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
```

### Pattern B: Deterministic Security Check Verification

```go
package assessment_test

import (
	"testing"
)

func TestEvaluateIDORBoundary(t *testing.T) {
	tests := []struct {
		name             string
		requestUser      string
		targetUser       string
		statusCode       int
		responseBody     string
		expectedVerified bool
		expectedSeverity string
	}{
		{
			name:             "legitimate access to own profile",
			requestUser:      "user_101",
			targetUser:       "user_101",
			statusCode:       200,
			responseBody:     `{"id": "user_101", "email": "alice@example.com"}`,
			expectedVerified: false, // expected behavior, not a vulnerability
			expectedSeverity: "",
		},
		{
			name:             "unauthorized access to other profile returns 403",
			requestUser:      "user_101",
			targetUser:       "user_102",
			statusCode:       403,
			responseBody:     `{"error": "Forbidden"}`,
			expectedVerified: false, // properly enforced access control
			expectedSeverity: "",
		},
		{
			name:             "idor confirmed: unauthorized profile data returned",
			requestUser:      "user_101",
			targetUser:       "user_102",
			statusCode:       200,
			responseBody:     `{"id": "user_102", "email": "bob@example.com", "ssn": "000-11-2222"}`,
			expectedVerified: true,
			expectedSeverity: "high",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Evaluates deterministic authorization logic against mock response
			verified, severity := evaluateMockResponse(tc.requestUser, tc.targetUser, tc.statusCode, tc.responseBody)
			if verified != tc.expectedVerified || severity != tc.expectedSeverity {
				t.Errorf("evaluateIDOR = (%v, %q); want (%v, %q)", verified, severity, tc.expectedVerified, tc.expectedSeverity)
			}
		})
	}
}

func evaluateMockResponse(reqUser, targetUser string, code int, body string) (bool, string) {
	if reqUser != targetUser && code == 200 && len(body) > 0 {
		return true, "high"
	}
	return false, ""
}
```

---

## 5. Pre-Commit Verification Checklist

Before committing any code change or opening a pull request:

1. Backend tests pass: `go test ./...`
2. Race condition free: `go test -race ./...`
3. Static frontend compiles: `cd web && npm run build`
4. Code formatted: `gofmt -s -w .`
5. Clean repository status: No untracked cache files, binaries, or databases.
