# BHF (Bug Hunting Framework) - Testing Strategy and Guidelines

This document outlines the testing architecture, testing principles, suite structure, and implementation standards for BHF. It serves as the definitive reference for maintaining high test coverage and preventing regressions without introducing brittle or bloated test code.

---

## 1. Testing Philosophy

- **Deterministic and Fast**: Tests must execute in milliseconds without external network calls, real port scans, or live web requests against target domains.
- **Isolated State**: Every database test must operate against a dedicated temporary SQLite database (`t.TempDir()`) or an in-memory instance (`:memory:`), ensuring complete test isolation and zero cross-test interference.
- **Controlled Growth (No Exponential Bloat)**:
  - Test business logic, parsing accuracy, scoping boundaries, and state transitions.
  - Avoid writing hundreds of identical redundant tests for simple struct getters.
  - Prefer clean, table-driven tests (`[]struct{ name, input, expected }`) that test multiple scenarios within a single, readable test function.
- **Zero Flakiness**: Tests must not rely on `time.Sleep()` or arbitrary delays. Synchronize goroutines using channels, wait groups (`sync.WaitGroup`), or context cancellations.
- **Zero Emojis**: Never use emojis in test function names, log statements, or assertion failure messages.

---

## 2. Test Architecture by Layer

### 2.1 Database Repositories (`internal/database/`)

Tests in this package verify SQLite migrations, CRUD operations, foreign key cascades, and complex analytical queries (such as asset prioritization and finding statistics).

**Key Rules**:
- Initialize a fresh database with `filepath.Join(t.TempDir(), "test.db")` or `:memory:`.
- Execute `MigrateSchema(db)` before testing repositories.
- Verify cascading deletes: deleting a target must automatically clean up associated scans, assets, findings, and logs.

**Reference Implementations**:
- `internal/database/asset_repo_test.go`: Verifies asset upserts, conflict resolution (`UNIQUE(target_id, host, port)`), and priority score calculation.
- `internal/database/finding_repo_test.go`: Verifies finding insertions, severity ordering, and aggregate count statistics.

### 2.2 Security Tools and Output Parsers (`internal/tools/`)

Security tools wrap external CLI binaries. Tool tests must NEVER execute the real binaries or send live packets. Instead, test the parsing engine and scope filtering logic:

1. **Scoping and Domain Sanitization**:
   - `CleanDomain()`: Verify stripping of `*.`, `https://`, `http://`, ports, and trailing slashes.
   - `IsInScope()`: Verify that target subdomains are in scope while foreign domains, third-party CDNs, and out-of-scope targets are strictly rejected.
2. **Parser Unit Tests**:
   - Test `Tool.Run(execCtx, rawOutput)` by supplying fixed strings representing CLI output or JSON lines.
   - Test error handling when output files or stdout lines contain malformed JSON or empty strings.

### 2.3 HTTP API Layer (`internal/api/`)

API handlers are tested using Go's standard `net/http/httptest` package:
- Use `httptest.NewRecorder()` to capture HTTP responses.
- Construct `http.Request` instances via `httptest.NewRequest()`.
- Inject repository instances backed by an in-memory SQLite database.
- Assert HTTP status codes, `Content-Type` headers, and JSON response payloads.

### 2.4 Model Context Protocol (`internal/mcp/`)

Tests in `internal/mcp/server_test.go` verify JSON-RPC 2.0 protocol conformance:
- `initialize`: Verify protocol version and server capabilities.
- `tools/list`: Verify that all registered tools and their input schemas are reported.
- `tools/call`: Verify tool execution (e.g., `list_targets`, `get_target_stats`) and error responses on invalid parameters.

### 2.5 Configuration Loader (`internal/config/`)

Tests in `internal/config/loader_test.go` verify:
- Default configuration values when environment variables are unset.
- Correct parsing and overrides when environment variables (`BHF_PORT`, `BHF_HOST`, `BHF_DATA_DIR`) are supplied.

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

## 4. Standard Table-Driven Test Pattern

When adding new tests in Go, follow this standard table-driven pattern:

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

---

## 5. Pre-Commit Verification Checklist

Before committing any code change or opening a pull request:

1. Backend tests pass: `go test ./...`
2. Race condition free: `go test -race ./...`
3. Static frontend compiles: `cd web && npm run build`
4. Code formatted: `gofmt -s -w .`
5. Clean repository status: No untracked cache files, binaries, or databases.
