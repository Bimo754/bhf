# Changelog

All notable changes to the BHF (Bug Hunting Framework) project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Created foundational architectural and engineering documentation:
  - `RULES.md`: Core engineering commandments, Single Responsibility Principle, UI aesthetics, traffic safety, and zero-guesswork debugging rules.
  - `ARCHITECTURE.md`: Complete system architecture, component diagram, directory roles, database schema, and MCP integration blueprint.
  - `TESTING.md`: Testing strategy, test packages, table-driven patterns, and pre-commit verification checklist.
  - `WORKFLOW.md`: 7-step feature addition lifecycle, Python tool isolation runbook, debugging protocol, and session starter prompt.
  - `CHANGELOG.md`: Progress tracking and release history.
- Added Model Context Protocol (MCP) server supporting stdio and HTTP (`POST /api/mcp`) for AI agent integration.
- Added Vulnerability Findings pipeline and dedicated UI view (`FindingsView.tsx`) with severity filtering, reproduction commands, and HackerOne report generation.
- Added Asset Prioritization scoring engine calculating priority scores based on tech stack, sensitive keywords, and non-standard ports.
- Added tool integrations for `dnsx`, `naabu`, and `theharvester` with isolated virtual environment execution.
- Added Phase 1 core reconnaissance and vulnerability tools:
  - `alterx`: Subdomain permutation and pattern generator with automatic scope verification.
  - `subzy`: Subdomain takeover detector checking CNAME records against vulnerable services.
  - `asnmap`: Network range, ASN, and CIDR mapping tool.
- Added Phase 2 HTTP probing and origin discovery tools:
  - `tlsx`: TLS/SSL certificate grabber, Subject Alternative Name (SAN) extractor, and JARM fingerprinting.
  - `wafw00f`: Web Application Firewall (WAF) detector and fingerprinting utility with uv isolation.
- Added Phase 3 URL deduplication and JavaScript mining tools:
  - `uro`: URL decluttering and deduplication filter for cleaning crawler archives with uv isolation.
  - `jsluice`: Tree-sitter AST JavaScript parser for extracting hidden API routes and leaked secrets.
- Added Phase 4 content discovery and parameter fuzzing tools:
  - `arjun`: HTTP parameter discovery detecting hidden GET, POST, and JSON parameters with uv isolation.
  - `nomore403`: HTTP 401/403 forbidden bypass engine utilizing header injection and path mutations.
- Added Phase 5 specialized vulnerability assessment and cloud reconnaissance tools:
  - `corsy`: CORS misconfiguration scanner detecting arbitrary origin reflection with credentials.
  - `s3scanner`: Cloud storage bucket reconnaissance scanner for public Amazon S3, Google Cloud, and Azure Blobs.
- Added Phase 6 visual reconnaissance and screenshot triage:
  - `gowitness`: Headless browser screenshot utility for rapid visual triage of live web applications.

### Changed
- Standardized UI to developer-grade GitHub Dark theme (`#0d1117`) with zero AI glowing neon hues and zero emojis.
- Enforced strict Single Responsibility Principle across all Go backend and TypeScript frontend files.
- Enhanced `.gitignore` with comprehensive patterns for Python bytecode, virtual environments, logs, and build artifacts.
- Removed all emojis from project documentation and user interface components.

### Fixed
- Fixed process cancellation handling to ensure all discovered assets and findings are parsed and preserved in SQLite when a scan is halted.
- Fixed text selection contrast with high-contrast styling (`#1f6feb` background, `#ffffff` text).
- Replaced native HTML form controls with custom dark popovers (`ScopeDropdown`).
- Resolved Corsy tool packaging in runner container via dedicated git clone and isolated `uv` virtual environment.
- Corrected Subzy repository import path to `github.com/PentestPad/subzy` in tools-builder container stage.

### Tested
- Database integration test suite (`asset_repo_test.go`, `finding_repo_test.go`).
- MCP protocol test suite (`server_test.go`).
