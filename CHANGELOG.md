# Changelog

All notable changes to the BHF (Bug Hunting Framework) project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Implemented **Eye-Comfort Dark Mode & Visual De-Cluttering Overhaul**:
  - Re-engineered the color palette into an eye-comfortable deep dark theme: `#080c14` canvas, `#0e1420` surface containers, soft hairline borders `rgba(255, 255, 255, 0.06)`, and muted text typography (`#f1f5f9` headings, `#94a3b8` body, `#64748b` subtle).
  - Eliminated button clutter across all primary views:
    - `ToolsView.tsx`: Reduced 50+ scattered tool and stage buttons into a sleek **Pipeline Command Center** with a single primary run action, integrated stage filter dropdown, and clean, quiet tool rows with subtle status dots (`sky-400` running, `emerald-400` completed, `rose-400` failed, `slate-600` ready).
    - `NetworkSecurityView.tsx`: Replaced 4 separate profile buttons and massive neon alert cards with a single unified search & scan row with integrated native port profile selection and sleek, contextual surface discovery notices.
    - `Navbar.tsx`: Replaced high-contrast saturated pill badges with quiet, minimal segmented pill navigation with micro mono counts.
    - `TrafficView.tsx`: Reorganized multi-button toolbar into clean primary `Auto-Explore Site` and `Run Security Audit` actions with quiet dropdown filters and subdued syntax formatting in `#080c14`.
    - `Button.tsx`: Modernized all buttons with refined Indigo accents (`#4f46e5`), soft translucent secondary surfaces, and subtle interactive active states.
    - `Toast.tsx`: Redesigned notification popups with soft dark slate surfaces and clear typography.
- Implemented **Unified Multi-Discipline UI Architecture & Cross-Discipline Dynamic Handoffs**:
  - Re-architected primary navigation into five distinct security disciplines: **Web Security**, **Network Security**, **Active Directory**, **Findings & Evidence**, and **Live Terminal**.
  - Grouped all web reconnaissance, asset inventory, visual screenshots, and traffic intelligence into a unified container (`WebSecurityView.tsx`) with fast sub-navigation (`[ Methodology & Tools | Discovered Assets | App Intel & Traffic Explorer ]`).
  - Implemented `NetworkSecurityView.tsx`: CIDR subnet and IP range probing, customizable port profiles (Core, Web, AD, Full), live host inventory, protocol fingerprinting, dangerous exposure alerting (unauthenticated Redis, Telnet, FTP, Elasticsearch, MongoDB), and dynamic 1-click discipline handoffs.
  - Implemented `ActiveDirectoryView.tsx`: Domain Controller detection, LDAP vs LDAPS transport security auditing, Kerberos KDC verification, and SMB signing analysis with direct finding generation.
  - Implemented seamless Cross-Discipline Handoffs: Web ports discovered during network scanning trigger 1-click handoff to Web Security with adjusted scope; AD ports trigger 1-click handoff to Active Directory auditing.
  - Updated `Navbar.tsx` and `App.tsx` with live discipline inventory badges and responsive state management.
- Implemented **Autonomous Application Explorer & Dynamic DOM Crawler (`AutoExplorer`)**:
  - Implemented `internal/appintel/auto_explorer.go`: Autonomous dual-mode explorer navigating target web applications, clicking interactive DOM elements, submitting forms with benign audit values, and capturing all underlying network flows without requiring manual proxy interaction.
  - Chromium CDP Integration: Automatically binds headless Chromium to an internal proxy, opens pages, evaluates DOM interaction scripts, executes button/tab clicks, fills inputs, scrolls pages, and intercepts 100% of network traffic into `traffic_records`.
  - HTTP Spider Fallback: Pure-Go crawler parsing HTML, discovering links and forms, submitting POST/GET requests with `RecordingTransport` for environments without Chromium.
  - Added REST API endpoints (`POST /api/appintel/target/{id}/auto-explore` and `GET /status`).
  - Added MCP tool `auto_explore_target` for autonomous AI agent site discovery.
  - Added "Auto-Explore Site" one-click action and live progress summary banner in `TrafficView.tsx`.
- Implemented **Interactive Application Intelligence & Traffic Inspection UI (`TrafficView.tsx`)**:
  - Added dedicated "App Intel" tab to the primary navigation bar (`Navbar.tsx` and `App.tsx`).
  - Implemented category filtering chips (`auth`, `admin`, `api`, `upload`, `search`, `ugc`, `general`) and HTTP method filtering (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`).
  - Implemented real-time category statistics counter banner bound to `trafficService.getStats()`.
  - Built interactive split-pane inspector displaying raw/formatted request headers, syntax-formatted request bodies, response headers, response bodies, extracted parameter graphs, and transaction metadata.
  - Added 1-click executable `cURL` reproduction command copying with parameter escaping for reproducible testing.
- Implemented **Phase 2: Traffic Categorization, Parameter Graph & Sensitive Data Protection**:
  - Implemented `internal/appintel/categorizer.go`: Multi-heuristic classification of dynamic workflows (`auth`, `registration`, `account_recovery`, `upload`, `search`, `ugc`, `admin`, `api`).
  - Implemented `internal/appintel/parameter_map.go`: Aggregated parameter graph tracking user-controlled parameters across query, JSON, and form bodies with type inference (`uuid`, `integer`, `email`, `token`) and automated IDOR candidate detection.
  - Implemented `internal/appintel/session.go`: Sensitive data redactor scrubbing passwords, JWTs, Bearer tokens, cookies, API keys, and credit cards for safe UI presentation and logging.
  - Unit test suite in `internal/appintel/appintel_test.go`.
- Implemented **Phase 3: Deterministic Security Assessment Engine & Evidence Verification**:
  - Implemented `internal/assessment/rule.go`: Verification rule interface, confidence scoring (`CONFIRMED`, `HIGH`, `MEDIUM`, `INFORMATIONAL`), and finding mapping.
  - Implemented `internal/assessment/evidence.go`: RFC 7230 raw request/response serializer and copy-pasteable `curl` reproduction command compiler.
  - Implemented `internal/assessment/auth_check.go`: Credential query parameter exposure and missing session cookie defensive flags (`HttpOnly`, `Secure`).
  - Implemented `internal/assessment/idor_check.go`: Controlled numeric identifier substitution and broken object-level authorization verification.
  - Implemented `internal/assessment/input_check.go`: Context-aware canary reflection and unescaped HTML tag verification.
  - Implemented `internal/assessment/upload_check.go`: Active script and SVG dangerous file extension acceptance audit.
  - Implemented `internal/assessment/engine.go`: Multi-rule assessment coordinator with automated finding persistence to SQLite.
  - Unit test suite in `internal/assessment/assessment_test.go`.
- Implemented **Phase 4: Hybrid AI Assessment Planner & Model Context Protocol Suite**:
  - Implemented `internal/ai/client.go`: Multi-provider LLM interface supporting Gemini, OpenAI, Ollama, and deterministic Mock testing.
  - Implemented `internal/ai/traffic_analyzer.go`: Structured prompt compiler and proposal parser for traffic graphs.
  - Implemented `internal/ai/planner.go`: Coordinated planner with deterministic validation gates.
  - Expanded `internal/mcp/server.go`: Added `query_traffic` and `get_parameter_graph` JSON-RPC tools for autonomous agent interaction.
  - Unit test suites in `internal/ai/ai_test.go` and `internal/mcp/server_test.go`.
- Implemented **Phase 5 & 6: Network Security Subsystem & Active Directory Mapping**:
  - Created domain entities `internal/models/host.go` and `internal/models/service.go`.
  - Added `hosts` and `services` SQLite tables with foreign keys and indexes in `internal/database/schema.go`.
  - Implemented `internal/database/host_repo.go` and `internal/database/service_repo.go` with automated web service handoff queries.
  - Implemented `internal/network/discovery.go`: Subnet CIDR expansion and TCP SYN/connect host probing.
  - Implemented `internal/network/service_probe.go`: Protocol fingerprinting (SSH, HTTP, Redis) and automated handoff to web recon (`HandoffWebServices`).
  - Implemented `internal/network/ad_mapper.go`: Non-destructive Active Directory Domain Controller detection and unencrypted LDAP auditing.
  - Implemented `internal/network/misconfig.go`: Network exposure auditing for unauthenticated Redis, Telnet, FTP, Elasticsearch, and MongoDB.
  - Unit test suites in `internal/database/network_repo_test.go` and `internal/network/network_test.go`.
- Implemented **Phase 7: Codebase Cleanup & Tool Retirement**:
  - Removed deprecated `internal/tools/nuclei.go` and `internal/tools/nomore403.go` source files.
  - Cleaned tool registrations in `cmd/bhf-server/main.go` and `internal/tools/tools_test.go`.
  - Cleaned `docker/Dockerfile` by removing Nuclei/nomore403 binaries and payload cloning stages.
  - Updated `web/src/views/ToolsView.tsx`, `RunConfigModal.tsx`, and `VisualGallery.tsx` removing Nuclei/403 bypass references and rebuilding production frontend bundles.
- Implemented **Phase 1: Interactive Application Intelligence & Traffic Interception Foundation**:
  - Created `internal/models/traffic.go` defining the `TrafficRecord` domain entity for full request/response transactions.
  - Added canonical `traffic_records` SQLite table, compound indexes, and cascading foreign key constraints in `internal/database/schema.go`.
  - Implemented `internal/database/traffic_repo.go` supporting paginated queries, category/method filtering, search, and category metrics with complete unit tests in `traffic_repo_test.go`.
  - Implemented `internal/appintel/interceptor.go` providing `RecordingTransport` (`http.RoundTripper`) and standalone `ProxyServer` with automatic in-scope enforcement (`tools.IsInScope`), header/body capture, and initial category inference.
  - Implemented `internal/appintel/browser.go` managing isolated headless Chromium sessions with process group isolation (`Setpgid`), temp profile directories, and graceful cancellation.
  - Implemented `internal/api/traffic_handler.go` and registered REST endpoints (`GET /api/traffic/target/{targetID}`, `/stats`, `/{id}`) in `internal/api/router.go` and `cmd/bhf-server/main.go`.
  - Added TypeScript interfaces in `web/src/types/traffic.ts` and network client in `web/src/services/trafficService.ts`.
- Added architecture and engineering specifications for **Interactive Application Intelligence** (`internal/appintel/`), introducing isolated browser session exploration via CDP, full HTTP transaction interception, traffic categorization (authentication, registration, reset, search, upload, admin, API), and parameter graph mapping.
- Added architecture and engineering specifications for the **Deterministic Security Assessment Engine** (`internal/assessment/`), shifting BHF from speculative scanner alert output to verified, reproducible security testing with explicit confidence levels, raw request/response proof, and executable `curl` reproductions.
- Added architecture and engineering specifications for the **Hybrid AI Assessment Orchestrator** (`internal/ai/`), coupling multi-provider AI agents for application comprehension, traffic analysis, and test planning with deterministic validation gates.
- Added architecture and engineering specifications for the **Network Security Subsystem** (`internal/network/`), detailing subnet discovery, host identification, service fingerprinting, Active Directory/identity assessment, and unified asset/finding models.
- Added architectural documentation for the `endpoints` table, `endpoint_repo.go`, and `endpoint_handler.go`, synchronizing documentation with the existing route separation implementation.
- Added "Skip Tool" feature allowing operators to skip any active tool during an automated sequence (full pipeline or stage sequence) from both the Tools and Terminal views without aborting the rest of the pipeline.
- Integrated native visual screenshot capture into `httpx` (`-screenshot`, `-system-chrome`, `-srd`, `-esb`, `-ehb`), parsing output paths and binding screenshots directly to asset inventory.
- Enhanced `VisualGallery.tsx` with default "Screenshots only" display, canonical host deduplication (collapsing 301/302 redirects into destination asset), and 4xx/5xx error suppression.

### Changed
- Realigned platform architecture from a simple external reconnaissance tool runner into a unified, evidence-driven security assessment and penetration testing platform.
- Replaced legacy tool runner pipeline with the 8-stage Progressive Security Assessment Methodology across Reconnaissance, Application Intelligence, Deterministic Assessment, and Network Security.
- Updated `ARCHITECTURE.md`, `RULES.md`, `TESTING.md`, `WORKFLOW.md`, `personal.md`, and `README.md` to reflect the core philosophy: *Reliability over quantity. Evidence over assumptions. Verification over speculation.*
- Standardized UI to developer-grade GitHub Dark theme (`#0d1117`) with zero AI glowing neon hues and zero emojis.
- Enforced strict Single Responsibility Principle across all Go backend and TypeScript frontend files.
- Enhanced `.gitignore` with comprehensive patterns for Python bytecode, virtual environments, logs, and build artifacts.
- Removed all emojis from project documentation and user interface components.
- Added view mode toggle `[ Table View | Visual Gallery | Sitemap Tree ]` to `AssetsView.tsx`.
- Retired `BUG-002` (Nuclei scan duration) and closed `BUG-006` (Nuclei missing templates) following tool deprecation.

### Removed
- Deprecated and removed `nuclei` template scanner from system architecture, methodology pipeline, and active bug tracking in favor of native, deterministic verification modules.
- Deprecated and removed `nomore403` blind header/path tampering from system architecture and roadmap (`FEAT-006`), replacing it with deterministic authorization and access control boundary checks.
- Purged all legacy external framework references across all project documentation and engineering rules.
- Deprecated and removed `alterx` (subdomain permutations) and `asnmap` (network range/ASN mapping) tools from the framework backend, tool registry, Docker container build, pipeline stages, and test suites per operator request (`BUG-007`).

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
