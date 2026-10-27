# BHF - Personal Tracking and Operator Log

This document is your personal scratchpad and tracking dashboard for recording bugs, feature ideas, and operational notes while actively using the framework on targets.

---

## 1. Active Bug Backlog

Track defects, crashes, parser anomalies, or UI issues discovered during live execution.

| ID | Date | Recon Stage | Severity | Summary | Status |
|---|---|---|---|---|---|
| BUG-001 | 2026-10-26 | Core | Low | Terminal autoscroll pauses when clicking inside container | Closed |
| BUG-002 | 2026-10-26 | Vulnerability Assessment | High | Nuclei scan duration takes an excessive amount of time to complete against multiple targets | Open |
| BUG-003 | 2026-10-26 | Discovered Assets / UI | Medium | Discovered Assets pagination shows empty intermediate pages when filtering by Live Web | Open |

*Status Values*: `Open`, `Investigating`, `In Progress`, `Resolved`, `Closed`

### Detailed Bug Descriptions

#### [BUG-002] Nuclei Scan Excessive Execution Time
- **Date**: 2026-10-26
- **Recon Stage**: Vulnerability Assessment and Verification
- **Tool Involved**: `nuclei`
- **Symptom**: When running Nuclei against discovered targets or domains, the scan requires an impractically long time to complete (potentially stalling for hours or days).
- **Suspected Root Cause**:
  - Scanning broad default template suites across multiple endpoints without template filtering.
  - Conservative concurrency (`-c 25`) and rate limits (`-rl 150`) combined with high timeout values on unresponsive endpoints.
  - Lack of targeted technology-based template filtering (e.g., matching detected technologies from `httpx` to specific Nuclei tags rather than running all templates).
- **Target Resolution**:
  - Implement intelligent template scoping (run tags matching detected tech stacks).
  - Provide severity selection and category presets (e.g., `exposures`, `misconfigurations`, `cves`).
  - Optimize concurrency and timeout thresholds for responsive execution.

#### [BUG-003] Discovered Assets Pagination and Filter Discrepancy
- **Date**: 2026-10-26
- **Recon Stage**: Discovered Assets / Web UI
- **Tool Involved**: `AssetTable.tsx`, `AssetHandler.go`
- **Symptom**: When selecting the "Live Web" filter chip on the Discovered Assets page, the table displays 16 pages, but clicking next shows multiple empty pages before reaching pages with matching entries.
- **Suspected Root Cause**:
  - The frontend applies client-side filtering (`filteredAssets = data.assets.filter(...)`) against a paginated slice returned by the backend (`LIMIT 50 OFFSET X`).
  - If page 1 contains 50 subdomains with no live status code, the client-side filter yields 0 rows, rendering a blank table for that page while the pagination counter still reflects total database assets.
- **Target Resolution**:
  - Move filter chips (`live`, `2xx`, `3xx`, `errors`, `priority`) to server-side query parameters in `GET /api/assets/target/{targetID}?filter=live`.
  - Let SQLite calculate total matching records and apply `WHERE` clauses prior to `LIMIT` and `OFFSET`.

---

## 2. Feature & Tool Wishlist

Track capabilities, tool integrations, and workflow improvements that increase bug bounty yield.

| ID | Date | Methodology Stage | Proposed Capability / Tool | Hunter ROI / Justification | Status |
|---|---|---|---|---|---|
| FEAT-001 | 2026-10-26 | Live HTTP Probing | TLS certificate SAN extraction | Finds hidden dev/origin hostnames on single IP addresses | Backlog |
| FEAT-002 | 2026-10-26 | Content Discovery | Custom header injection in Ffuf (Auth / Cookie) | Enables authenticated endpoint brute forcing | Planned |
| FEAT-003 | 2026-10-26 | Crawling & Endpoints | Dedicated Endpoints table & UI tab | Separates crawled web routes from host assets; enables route search | Backlog |
| FEAT-004 | 2026-10-26 | Execution Engine | Real-time streaming ingestion | Pipes stdout lines directly into DB and WS live during execution | Backlog |
| FEAT-005 | 2026-10-26 | Crawling & Endpoints | JavaScript route & secret mining (jsluice) | Uncovers hidden API routes and leaked tokens from crawled JS | Backlog |
| FEAT-006 | 2026-10-26 | HTTP Probing / Vuln | 403/401 bypass engine (nomore403) | Automates header tampering and path mutations on forbidden endpoints | Backlog |
| FEAT-007 | 2026-10-26 | Content Discovery | Parameter discovery (arjun / x8) | Discovers hidden query and body parameters for injection testing | Backlog |
| FEAT-008 | 2026-10-26 | Live HTTP Probing | Visual reconnaissance gallery (gowitness / httpx) | Automated web screenshots and UI visual triage gallery grid | Backlog |
| FEAT-009 | 2026-10-26 | Progressive Pipeline | Continuous recon delta engine | Historical snapshots and diffing ([NEW], [CHANGED], [REMOVED]) | Backlog |
| FEAT-010 | 2026-10-26 | Workflow Orchestration| Automated multi-stage pipeline chaining | 1-click chained scans across stages with operator approval | Backlog |
| FEAT-011 | 2026-10-27 | Execution Engine | Subprocess group isolation & emergency kill switch | Prevents orphaned processes; allows 1-click abort of all running scans | Backlog |
| FEAT-012 | 2026-10-27 | UI & Notifications | Persistent notification drawer & OS desktop bread alerts | Notifies hunter when scans complete/stop while working in other windows | Backlog |

*Status Values*: `Idea`, `Planned`, `In Progress`, `Implemented`, `Rejected`

### Detailed Feature Specifications

#### [FEAT-003] Dedicated Endpoints Table and UI View
- **Methodology Stage**: Crawling and Endpoint Discovery
- **Hunter Justification**: Host-level assets (subdomains, IPs) have different lifecycles and metadata than endpoint URLs (paths, HTTP methods, parameters). Storing thousands of crawled URLs inside `assets` clutters host triage.
- **Architecture**:
  - SQLite table `endpoints` (`id`, `target_id`, `asset_id`, `url`, `path`, `method`, `status_code`, `params`, `source_tool`, `created_at`).
  - Separate crawled web paths (from Katana and GAU) from the host-level `assets` table.
  - Dedicated "Endpoints" tab in UI with search, path filtering, and export.

#### [FEAT-004] Real-Time Streaming Ingestion
- **Methodology Stage**: Execution Engine and Streaming
- **Hunter Justification**: Eliminates blind waiting during long-running crawler or fuzzing jobs. Discovered assets appear live in inventory while the process is actively executing.
- **Architecture**:
  - Refactor `internal/engine/runner.go` and `scan_handler.go` to pipe stdout lines into streaming tool parsers as lines arrive.
  - Persist records to SQLite and broadcast via WebSocket immediately.

#### [FEAT-005] JavaScript Route and Secret Mining (`jsluice`)
- **Methodology Stage**: Crawling and Endpoint Discovery
- **Hunter Justification**: Modern single-page applications leak API routes, internal endpoints, authorization tokens, and API keys inside compiled JavaScript bundles.
- **Architecture**:
  - Add `jsluice` binary to Dockerfile.
  - Create `internal/tools/jsluice.go` to parse discovered JS files from Katana/GAU.

#### [FEAT-006] 403/401 Forbidden Bypass Engine (`nomore403`)
- **Methodology Stage**: Live HTTP Probing / Vulnerability Assessment
- **Hunter Justification**: `403 Forbidden` and `401 Unauthorized` endpoints are primary targets for authentication bypass and administrative access vulnerabilities.
- **Architecture**:
  - Add `nomore403` to Dockerfile.
  - Create `internal/tools/nomore403.go` targeting live 401/403 assets with header manipulation and path mutations.

#### [FEAT-007] Parameter Discovery (`arjun` or `x8`)
- **Methodology Stage**: Content Discovery and Fuzzing
- **Hunter Justification**: Hidden GET and POST parameters frequently bypass validation, leading to SSRF, SQLi, IDOR, and XSS.
- **Architecture**:
  - Install `arjun` via isolated `uv tool` or `x8` binary in Docker runner.
  - Create `internal/tools/arjun.go` to discover hidden parameters on live web endpoints.

#### [FEAT-008] Visual Reconnaissance Gallery Grid
- **Methodology Stage**: Live HTTP Probing
- **Hunter Justification**: Rapid visual scanning of hundreds of web endpoints allows hunters to spot developer panels, login portals, default installation pages, and internal dashboards within seconds.
- **Architecture**:
  - Use `gowitness` or `httpx -ss` to generate screenshots saved to `/app/data/screenshots/`.
  - Add a Visual Gallery view in the UI rendering screenshot cards with click-to-browse.

#### [FEAT-009] Continuous Recon Delta Engine
- **Methodology Stage**: Progressive Pipeline & Analytics
- **Hunter Justification**: Bounties are often found on newly deployed subdomains and altered endpoints before other researchers notice them.
- **Architecture**:
  - Store historical scan snapshots upon completion.
  - Compute actionable diffs against previous runs:
    - `[NEW]`: Newly resolved subdomains, ports, or URLs.
    - `[CHANGED]`: HTTP status code changes (e.g. 403 -> 200) or response size shifts.
    - `[REMOVED]`: Inactive or dead assets.
  - Add visual delta badges and a "Changes Since Last Scan" filter chip in the UI.

#### [FEAT-010] Automated Multi-Stage Pipeline Chaining
- **Methodology Stage**: Workflow Orchestration
- **Hunter Justification**: Eliminates repetitive manual execution while preserving 100% operator approval.
- **Architecture**:
  - Pipeline coordinator linking stages sequentially:
    `Subdomain Enumeration -> DNS Resolution -> Port Scanning -> HTTP Probing -> Crawling -> Vulnerability Assessment / JS Mining`.
  - Visual stage progression indicator in UI with single-click start and stop.

#### [FEAT-011] Subprocess Group Termination & Emergency Kill Switch
- **Methodology Stage**: Execution Engine & Process Management
- **Hunter Justification**: Complex offensive tools often spawn child processes or background threads (e.g. headless browser instances in Katana, subprocess workers in Python tools). When an operator cancels a scan, standard Go context cancellation can leave orphaned child processes continuing to flood targets or consume CPU. Process group isolation and a global kill switch ensure clean termination.
- **Architecture**:
  - Backend: In `internal/engine/runner.go`, set `cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}` and terminate the entire process group (`syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)`).
  - Add `POST /api/scans/kill-all` in `internal/api/scan_handler.go` and `internal/engine/dispatcher.go` to abort all active scans across targets in one click.
  - Frontend: Add an "Emergency Abort / Kill Running" button in the Navbar or Tools view when scans are active.

#### [FEAT-012] Persistent Notification Drawer & OS Desktop Bread Notifications
- **Methodology Stage**: UI & Operator Notifications
- **Hunter Justification**: Long reconnaissance scans run while hunters work in other windows (Burp Suite, terminal, notes). System desktop notifications and an in-app persistent notification history ensure the operator never misses a completed scan, failed process, or critical vulnerability discovery.
- **Architecture**:
  - Web Desktop Notifications: Integrate HTML5 Web Notifications API (`Notification.requestPermission()`) in `web/src/services/` to display native OS notification banners when scans finish, fail, or stop while the browser tab is inactive or minimized.
  - In-App Notification Center: Add an accessible notification drawer/tray in `web/src/components/` logging all scan completion/stop events and critical alerts with 1-click navigation to the scan console or discovered findings.

---

## 3. Operational Checklist: Testing the MCP Server

Use this testing checklist to verify the Model Context Protocol (MCP) server integration:

### Mode 1: HTTP API Verification (`POST /api/mcp`)
Ensure the BHF server is running (e.g., `make run` or inside Docker at `http://localhost:8080`):

- [ ] **Test 1: Initialize Protocol Handshake**
  ```bash
  curl -s -X POST http://localhost:8080/api/mcp \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | jq .
  ```
  *Expected*: Returns `protocolVersion: "2024-11-05"`, `capabilities.tools`, and `serverInfo.name: "bhf-mcp-server"`.

- [ ] **Test 2: List Available Tools**
  ```bash
  curl -s -X POST http://localhost:8080/api/mcp \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | jq .
  ```
  *Expected*: Returns JSON list of all 5 tools (`list_targets`, `list_assets`, `list_findings`, `get_target_stats`, `list_tools`) with input schemas.

- [ ] **Test 3: Call `list_targets`**
  ```bash
  curl -s -X POST http://localhost:8080/api/mcp \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_targets","arguments":{}}}' | jq .
  ```
  *Expected*: Returns registered targets array.

- [ ] **Test 4: Call `get_target_stats`**
  ```bash
  curl -s -X POST http://localhost:8080/api/mcp \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_target_stats","arguments":{"target_id":"YOUR_TARGET_UUID"}}}' | jq .
  ```
  *Expected*: Returns combined asset counts and vulnerability severity breakdown.

- [ ] **Test 5: Call `list_assets` and `list_findings`**
  ```bash
  curl -s -X POST http://localhost:8080/api/mcp \
    -H "Content-Type: application/json" \
    -d '{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"list_assets","arguments":{"target_id":"YOUR_TARGET_UUID"}}}' | jq .
  ```
  *Expected*: Returns paginated discovered assets array.

### Mode 2: Standard I/O Verification (`stdio`)

- [ ] **Test 6: Run Stdio Mode**
  ```bash
  echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./bin/bhf-server --mcp
  ```
  *Expected*: Outputs single JSON-RPC line containing the tool definitions without starting HTTP server.

---

## 4. Live Hunting Notes & Target Observations

Operator scratchpad for noting quirks, WAF behaviors, wordlist tweaks, or high-value findings during live engagements:

- **Target Notes**:
  - Example: Keep rate limit for `example.com` under 50 req/sec to avoid Akamai rate-limiting.
  - Example: `admin-api.target.com:8443` returned a 403 forbidden; test HTTP method tampering (`X-Original-URL`, `X-Rewrite-URL`, `GET /` vs `POST /`).
