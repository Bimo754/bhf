# Changelog

All notable changes to the BHF (Bug Hunting Framework) project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Added "Skip Tool" feature allowing operators to skip any active tool during an automated sequence (full pipeline or stage sequence) from both the Tools and Terminal views without aborting the rest of the pipeline.
- Integrated native visual screenshot capture into `httpx` (`-screenshot`, `-system-chrome`, `-srd`, `-esb`, `-ehb`), parsing output paths and binding screenshots directly to asset inventory.
- Enhanced `VisualGallery.tsx` with default "Screenshots only" display, canonical host deduplication (collapsing 301/302 redirects into destination asset), and 4xx/5xx error suppression.

### Fixed
- Fixed screenshot loading failures in BHF Visual Gallery and Discovered Assets table by:
  - Preserving nested relative paths (`screenshot/<host>/<hash>.png`) emitted by `httpx -srd` instead of truncating via `filepath.Base()`.
  - Adding fallback parsing of httpx's `index_screenshot.txt` when JSON output omits `screenshot_path`.
  - Replacing rigid static screenshot file server with an intelligent, traversal-safe `screenshotHandler` that directly matches relative paths and automatically falls back to filename hash discovery within target screenshot folders.
  - Enforcing `0755` directory and `0644` file permissions post-scan so screenshots are world-readable across Docker and host environments.
  - Updating existing database asset records to point directly to valid screenshot paths.
  - Adding `onError` image fallback in `VisualGallery.tsx` to handle network issues gracefully without broken image icons.
- Replaced unstable `gowitness` tool with ProjectDiscovery's native `httpx -ss` engine, resolving `BUG-008` (browser CDP crashes, `/dev/shm` exhaustion, silent capture drops, and placeholder card clutter).
- Removed deprecated `gowitness` across backend, Dockerfile, tests, and methodology pipeline.
- Fixed `nomore403` instant 37ms termination by attaching `os.DevNull` character device to `cmd.Stdin` in `internal/engine/runner.go` rather than an empty pipe, preventing the tool from mistaking non-interactive execution for empty piped stdin.
- Fixed frontend scan state desynchronization in `web/src/App.tsx` by adding active scan status reconciliation and periodic polling, ensuring the UI does not freeze on fast-finishing scans or missed WebSocket broadcasts.
- Ensured scans are cleanly unregistered from dispatcher via deferred call in `internal/api/scan_handler.go`.
- Created foundational architectural and engineering documentation:
  - `RULES.md`: Core engineering commandments, Single Responsibility Principle, Don't Repeat Yourself (DRY), Single Source of Truth (SSOT), UI aesthetics, traffic safety, and zero-guesswork debugging rules.
  - `ARCHITECTURE.md`: Complete system architecture, component diagram, directory roles, database schema, SRP & DRY principles, and MCP integration blueprint.
  - `TESTING.md`: Testing strategy, test packages, table-driven patterns, and pre-commit verification checklist.
  - `WORKFLOW.md`: 7-step feature addition lifecycle with DRY gates, Python tool isolation runbook, debugging protocol, and session starter prompt.
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
- Added Discovered Endpoints Schema Separation (`FEAT-017`):
  - Created dedicated `endpoints` SQLite table with foreign keys, compound unique constraints, and indexes.
  - Implemented `EndpointRepo` and `EndpointHandler` (`GET /api/endpoints/target/{targetID}`).
  - Routed crawler and fuzzer outputs (`katana`, `ffuf`, `gau`, `uro`, `jsluice`) to `endpoints` table, preventing host row URL overwrites.
- Added Subdomain Hierarchical Sitemap & Tree View (`FEAT-015`):
  - Interactive tree view (`SitemapTree.tsx`) grouping endpoints by subdomain with HTTP method tags, status code badges, and path search.
- Added Dedicated Visual Reconnaissance Gallery Tab (`FEAT-016`):
  - Visual gallery grid (`VisualGallery.tsx`) rendering screenshot thumbnail cards with lightbox image preview and direct Nuclei scan launch.
- Added Subdomain Permutation DNS Pipeline (`FEAT-014`):
  - Piped `alterx` permutation generation directly into active `dnsx` resolution, ensuring only valid resolving A and CNAME records are recorded.
- Added Custom Header & Cookie Injection for Ffuf (`FEAT-002`):
  - Configurable `-H "Header: value"` and `-b "cookie"` inputs in `RunConfigModal.tsx` for authenticated directory and API fuzzing.
- Added Dedicated Priority Score & Star Rating Column (`FEAT-013`):
  - Dedicated column with star ratings, tier badges, and 1-click server-side sorting (`priority_score`).
- Added Automated Methodology Stage Chaining & Progressive Pipeline Coordinator (`FEAT-010`):
  - Deterministic tool ordering in `ToolsView.tsx` with numbered step badges (`#01`, `#02`, `#03`...).
  - Stage-level sequential execution button ("Run Stage") to automatically run all tools in a category in order.
  - "Run Full Recon Pipeline" coordinator with optional crawl feedback loop re-probing newly discovered subdomains with `httpx` after crawling & JS mining before fuzzing and vulnerability scanning.
  - Thread-safe scan queue advancement and abort controls in `App.tsx`.

### Removed
- Deprecated and removed `alterx` (subdomain permutations) and `asnmap` (network range/ASN mapping) tools from the framework backend, tool registry, Docker container build, pipeline stages, and test suites per operator request (`BUG-007`).

### Changed
- Standardized UI to developer-grade GitHub Dark theme (`#0d1117`) with zero AI glowing neon hues and zero emojis.
- Enforced strict Single Responsibility Principle across all Go backend and TypeScript frontend files.
- Enhanced `.gitignore` with comprehensive patterns for Python bytecode, virtual environments, logs, and build artifacts.
- Removed all emojis from project documentation and user interface components.
- Added view mode toggle `[ Table View | Visual Gallery | Sitemap Tree ]` to `AssetsView.tsx`.
- Calibrated default `nuclei` CLI flags (`-s low,medium,high,critical`, `-tags cve,exposure,misconfig,takeover,vuln`, `-c 50`, `-rl 300`, `-timeout 5`, `-retries 1`, `-mhe 3`, `-stats`, `-si 10`) to eliminate scan stalling (`BUG-002`).

### Fixed
- Fixed process cancellation handling to ensure all discovered assets and findings are parsed and preserved in SQLite when a scan is halted.
- Fixed text selection contrast with high-contrast styling (`#1f6feb` background, `#ffffff` text).
- Replaced native HTML form controls with custom dark popovers (`ScopeDropdown`).
- Resolved Corsy tool packaging in runner container via dedicated git clone and isolated `uv` virtual environment.
- Corrected Subzy repository import path to `github.com/PentestPad/subzy` in tools-builder container stage.
- Resolved jsluice module path casing (`github.com/BishopFox/jsluice`) and enabled CGO compilation for tree-sitter bindings.
- Fixed frontend runaway re-render and WebSocket reconnection storm in `App.tsx` by using refs for asynchronous callbacks and scoping effect dependencies to `activeScanId`.
- Fixed real-time unbuffered execution across Python utilities (`wafw00f`, `arjun`, `corsy`, `theHarvester`) by setting `PYTHONUNBUFFERED=1` in subprocess runners and Docker environments.
- Implemented `scanLinesOrCarriageReturn` in `runner.go` to split on carriage returns (`\r`), enabling live streaming for interactive terminal progress bars (`arjun`, `gowitness`, `katana`, `nuclei`, `ffuf`).
- Fixed silent output in `alterx` and `uro` by removing file output flags (`-o`) so generated entries stream immediately to stdout.
- Fixed `asnmap` hanging on interactive API key prompts by closing subprocess stdin and validating `PDCP_API_KEY` presence before execution.
- Fixed `tlsx` CLI failure caused by probe flag conflicts by removing `-san` and `-cn` flags when invoking `-jarm`.
- Fixed `gowitness` browser resolution and web integration by symlinking `/usr/bin/google-chrome`, tagging assets with screenshot URIs, serving screenshots via `/api/screenshots/*`, and adding an in-app screenshot modal in `AssetTable.tsx`.
- Fixed `jsluice` infinite hang by properly specifying `--resolve-paths` and passing target URLs as arguments while feeding EOF to stdin.
- Fixed `arjun` method syntax error by passing single `-m GET` parameter.
- Fixed `nomore403` flag syntax error by replacing `-max-rate 30` with `-m 30 --no-banner` and mounting payloads at `/opt/nomore403/payloads`.
- Fixed `corsy` "no valid urls to test" error by piping targets to stdin and supplying `-u targets[0]`.
- Fixed `s3scanner` command failure by eliminating the invalid `scan` subcommand and using `-bucket-file` with `-json`.
- Fixed `BUG-002` (Nuclei scan excessive execution duration) by optimizing CLI flags and adding category/severity presets in scan launcher.
- Fixed `BUG-003` (Discovered Assets pagination intermediate empty pages) by migrating live/status/priority filtering to server-side SQL queries.
- Fixed `BUG-004` (Scan stop failure toast message truncated) by wrapping text and expanding toast container dimensions.
- Recorded `BUG-006` (Nuclei missing templates error), `BUG-007` (Remove alterx integration), and `BUG-008` (Visual Gallery placeholders and duplicate cards) in `personal.md`.

### Tested
- Database integration test suite (`asset_repo_test.go`, `endpoint_repo_test.go`, `finding_repo_test.go`).
- HTTP API integration test suite (`api_test.go`).
- MCP protocol test suite (`server_test.go`).
