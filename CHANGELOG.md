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

### Changed
- Standardized UI to developer-grade GitHub Dark theme (`#0d1117`) with zero AI glowing neon hues and zero emojis.
- Enforced strict Single Responsibility Principle across all Go backend and TypeScript frontend files.
- Enhanced `.gitignore` with comprehensive patterns for Python bytecode, virtual environments, logs, and build artifacts.
- Removed all emojis from project documentation and user interface components.

### Fixed
- Fixed process cancellation handling to ensure all discovered assets and findings are parsed and preserved in SQLite when a scan is halted.
- Fixed text selection contrast with high-contrast styling (`#1f6feb` background, `#ffffff` text).
- Replaced native HTML form controls with custom dark popovers (`ScopeDropdown`).

### Tested
- Database integration test suite (`asset_repo_test.go`, `finding_repo_test.go`).
- MCP protocol test suite (`server_test.go`).
