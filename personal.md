# BHF - Personal Tracking and Operator Log

This document is your personal scratchpad and tracking dashboard for recording bugs, feature ideas, and operational notes while actively using the framework on targets.

---

## 1. Active Bug Backlog

Track defects, crashes, parser anomalies, or UI issues discovered during live execution.

| ID | Date | Recon Stage | Severity | Summary | Status |
|---|---|---|---|---|---|
| BUG-001 | 2026-10-26 | Core | Low | Terminal autoscroll pauses when clicking inside container | Closed |
| BUG-002 | 2026-10-26 | Vulnerability Assessment | High | Nuclei scan duration takes an excessive amount of time to complete against multiple targets | Resolved |
| BUG-003 | 2026-10-26 | Discovered Assets / UI | Medium | Discovered Assets pagination shows empty intermediate pages when filtering by Live Web | Resolved |
| BUG-004 | 2026-10-30 | UI / Notifications | Low | Scan stop failure toast message is truncated preventing operator from reading error details | Resolved |
| BUG-005 | 2026-10-30 | Content Discovery | High | Arjun stability probe raises unhandled AttributeError on string response | Resolved |
| BUG-006 | 2026-10-30 | Vulnerability Assessment | High | Nuclei scan fails immediately with 'no templates provided for scan' due to missing /root/nuclei-templates | Open |
| BUG-007 | 2026-10-30 | Tooling / Subdomain Enum | Low | Remove alterx and asnmap tool integrations from framework | Resolved |
| BUG-008 | 2026-10-30 | Live HTTP Probing / Visual | High | Screenshot engine instability (gowitness capture drops/crashes) and Visual Gallery placeholder clutter | Resolved |

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
- **Resolution**:
  - Calibrated default CLI flags in `internal/tools/nuclei.go` to `-s low,medium,high,critical`, `-tags cve,exposure,misconfig,takeover,vuln`, `-c 50`, `-rl 300`, `-timeout 5`, `-retries 1`, `-mhe 3`, `-stats`, `-si 10`.
  - Added severity selection and high-value category presets in `RunConfigModal.tsx` allowing fast surgical audits without full template suite scans.

#### [BUG-003] Discovered Assets Pagination and Filter Discrepancy
- **Date**: 2026-10-26
- **Recon Stage**: Discovered Assets / Web UI
- **Tool Involved**: `AssetTable.tsx`, `AssetHandler.go`, `AssetRepo.go`
- **Symptom**: When selecting the "Live Web" filter chip on the Discovered Assets page, the table displays 16 pages, but clicking next shows multiple empty pages before reaching pages with matching entries.
- **Suspected Root Cause**:
  - The frontend applies client-side filtering (`filteredAssets = data.assets.filter(...)`) against a paginated slice returned by the backend (`LIMIT 50 OFFSET X`).
  - If page 1 contains 50 subdomains with no live status code, the client-side filter yields 0 rows, rendering a blank table for that page while the pagination counter still reflects total database assets.
- **Resolution**:
  - Migrated filtering to server-side SQL queries (`ListByTarget` in `internal/database/asset_repo.go`) supporting `filter=live`, `2xx`, `3xx`, `errors`, `priority`, and `new`.
  - Updated `AssetHandler.go` to parse `filter`, `sort`, and `order` query parameters, ensuring pagination metadata reflects exactly the filtered subset with zero intermediate blank pages.

#### [BUG-004] Scan Stop Failure Truncated Toast Message
- **Date**: 2026-10-30
- **Recon Stage**: UI & Process Controls
- **Component**: `Toast.tsx`
- **Symptom**: When attempting to cancel or stop a scan that has already finished or failed, the UI renders an error toast notification whose message text is truncated with CSS ellipsis (`truncate`), hiding the full error explanation from the operator.
- **Resolution**: Replaced single-line CSS `truncate` with `break-words whitespace-pre-wrap max-h-48 overflow-y-auto` and widened `ToastContainer` to `max-w-md`, allowing multi-line error details to be read cleanly.

#### [BUG-005] Arjun Stability Probe Unhandled String Exception
- **Date**: 2026-10-30
- **Recon Stage**: Content Discovery & Parameter Fuzzing
- **Tool Involved**: `arjun`
- **Symptom**: Arjun terminates with `exit status 1` and traceback `AttributeError: 'str' object has no attribute 'status_code'` in `initialize` during target stability testing.
- **Root Cause**: On PyPI release `arjun==2.2.1`, when a target endpoint returns a connection error, SSL failure, or drops packets during stability probing, `requester()` returns an error message string instead of a Response object. Arjun immediately attempts to access `response_1.status_code`, triggering an unhandled Python exception.
- **Resolution**: Updated `docker/Dockerfile` to install Arjun from git master (`git+https://github.com/s0md3v/Arjun.git`), which incorporates upstream commit handling `if isinstance(response_1, str): return 'skipped'`.

#### [BUG-006] Nuclei Missing Templates Execution Failure
- **Date**: 2026-10-30
- **Recon Stage**: Vulnerability Assessment and Verification
- **Tool Involved**: `nuclei`
- **Symptom**: Nuclei scan terminates immediately with `exit status 1` and fatal error `no templates provided for scan`:
  ```
  WARNING: sonic/ast only supports (go1.17~1.26 and amd64 CPU) or (go1.20~1.26 and arm64 CPU), but your environment is not suitable and will fallback to encoding/json
  __ _
  ____ __ _______/ /__ (_)
  / __ \/ / / / ___/ / _ \/ /
  / / / / /_/ / /__/ / __/ /
  /_/ /_/\__,_/\___/_/\___/_/ v3.11.1
  projectdiscovery.io
  [INF] nuclei-templates are not installed, installing...
  [INF] Your current nuclei-templates are outdated. Latest is v10.5.1
  [ERR] Could not find template '/root/nuclei-templates': no templates found in path "/root/nuclei-templates"
  [INF] Current nuclei version: v3.11.1 (latest)
  [INF] Current nuclei-templates version: (outdated)
  [INF] Targets loaded for current scan: 1
  [INF] Scan completed in 176.589µs. No results found.
  [FTL] Could not run nuclei: no templates provided for scan
  process exited with error: exit status 1
  ```
- **Root Cause**:
  - The runtime environment / container does not have `nuclei-templates` pre-installed or cloned at the default lookup path (`/root/nuclei-templates`).
  - At runtime, Nuclei attempts on-the-fly template installation, but fails to locate or populate the template directory, causing it to abort immediately with exit status 1.
- **Target Resolution**: (Tracked, do not fix yet per operator request)
  - Pre-install or update `nuclei-templates` during Docker build (`RUN nuclei -update-templates` or clone into `/root/nuclei-templates`).
  - Ensure template path or `-update-templates` flag is available so Nuclei can locate valid templates before scanning.

#### [BUG-007] Remove "alterx" and "asnmap" Tool Integrations
- **Date**: 2026-10-30
- **Recon Stage**: Tooling & Subdomain Enumeration
- **Tools Involved**: `alterx`, `asnmap`
- **Symptom**: Operator requested deprecation and complete removal of the `alterx` and `asnmap` applications/tool integrations from the BHF platform.
- **Root Cause / Justification**: Subdomain permutation via `alterx` and ASN mapping via `asnmap` (which required external PDCP API keys) are no longer desired in the active recon suite.
- **Resolution**:
  - Removed `internal/tools/alterx.go` and `internal/tools/asnmap.go`.
  - Unregistered `alterx` and `asnmap` from tool registry in `cmd/bhf-server/main.go`.
  - Removed parser unit tests from `internal/tools/tools_test.go`.
  - Removed binary build instructions from `docker/Dockerfile`.
  - Removed tool references from `web/src/views/ToolsView.tsx` methodology stages and pipeline lists.
  - Updated documentation across `ARCHITECTURE.md` and `CHANGELOG.md`.

#### [BUG-008] Visual Gallery Screenshot Failures, Placeholders, and gowitness Engine Instability
- **Date**: 2026-10-30
- **Recon Stage**: Live HTTP Probing / Visual Reconnaissance
- **Component**: `VisualGallery.tsx`, `gowitness`, `httpx`
- **Severity**: High
- **Symptom**:
  - `gowitness` is highly unstable during live scanning: screenshots are frequently not taken at all, dropping targets silently without raising process errors.
  - Headless Chromium instances inside Docker/Linux crash, stall on complex SPAs, or exhaust shared memory (`/dev/shm`), causing entire batches of screenshots to fail.
  - The Visual Gallery renders numerous cards displaying placeholder "No Screenshot Available" because assets without screenshots are shown by default alongside screenshotted hosts.
  - Multiple duplicate cards exist for the same target service (e.g., 301/302 redirects from `http://` to `https://` or apex domain to `www`).
  - Many cards render failed capture states (e.g., 400 Bad Request, 502 Bad Gateway, SSL connection resets) where headless browser rendering failed or timed out.
- **Root Cause Analysis**:
  - **Engine Instability**: `gowitness` relies on external Chromium via Chrome DevTools Protocol (CDP). Inside container environments without `--no-sandbox`, `--disable-dev-shm-usage`, and `--disable-gpu`, headless Chromium is notoriously fragile and prone to silent `SIGTRAP`/`SIGBUS` crashes.
  - **Silent Drops & Incomplete Captures**: When `gowitness scan file` encounters unresponsive endpoints, cert mismatches, or hangs on heavy JavaScript bundles, it times out or skips targets without writing JPEG/PNG files to disk, yet `gowitness.go` still records asset entries.
  - **Frontend Filtering Gap**: `VisualGallery.tsx` filters only on `a.url || a.status_code`, without checking whether a valid screenshot URI exists in `a.technologies`.
  - **Redirect Duplication**: No client-side deduplication or canonical redirect folding (e.g. coalescing 301/302 redirect pairs into the destination 200 target).
- **Resolution**:
  - Replaced `gowitness` with ProjectDiscovery's native `httpx -ss -system-chrome -srd <screenshotDir> -esb -ehb` integration in `internal/tools/httpx.go`, unifying live HTTP probing and visual screenshots into a single, high-reliability tool.
  - Fully removed `gowitness` across backend (`cmd/bhf-server/main.go`), Dockerfile (`docker/Dockerfile`), tests (`internal/tools/tools_test.go`), and frontend (`ToolsView.tsx`).
  - Enhanced `VisualGallery.tsx` with default "Screenshots only" filtering, canonical host deduplication (collapsing 301/302 redirects into canonical 200 OK asset), and 4xx/5xx error suppression.

---

## 2. Feature & Tool Wishlist

Track capabilities, tool integrations, and workflow improvements that increase bug bounty yield.

| ID | Date | Methodology Stage | Proposed Capability / Tool | Hunter ROI / Justification | Status |
|---|---|---|---|---|---|
| FEAT-001 | 2026-10-26 | Live HTTP Probing | TLS certificate SAN extraction (tlsx) | Finds hidden dev/origin hostnames on single IP addresses | Implemented |
| FEAT-002 | 2026-10-26 | Content Discovery | Custom header injection in Ffuf (Auth / Cookie) | Enables authenticated endpoint brute forcing | Implemented |
| FEAT-003 | 2026-10-26 | Crawling & Endpoints | Dedicated Endpoints table & UI tab | Separates crawled web routes from host assets; enables route search | Implemented |
| FEAT-004 | 2026-10-26 | Execution Engine | Real-time streaming ingestion | Pipes stdout lines directly into DB and WS live during execution | Backlog |
| FEAT-005 | 2026-10-26 | Crawling & Endpoints | JavaScript route & secret mining (jsluice) | Uncovers hidden API routes and leaked tokens from crawled JS | Implemented |
| FEAT-006 | 2026-10-26 | HTTP Probing / Vuln | 403/401 bypass engine (nomore403) | Automates header tampering and path mutations on forbidden endpoints | Implemented |
| FEAT-007 | 2026-10-26 | Content Discovery | Parameter discovery (arjun) | Discovers hidden query and body parameters for injection testing | Implemented |
| FEAT-008 | 2026-10-26 | Live HTTP Probing | Visual reconnaissance gallery (gowitness) | Automated web screenshots and UI visual triage gallery grid | Implemented |
| FEAT-009 | 2026-10-26 | Progressive Pipeline | Continuous recon delta engine | Historical snapshots and diffing ([NEW], [CHANGED], [REMOVED]) | Backlog |
| FEAT-010 | 2026-10-26 | Workflow Orchestration| Automated multi-stage pipeline chaining | 1-click chained scans across stages with operator approval | Implemented |
| FEAT-011 | 2026-10-27 | Execution Engine | Subprocess group isolation & emergency kill switch | Prevents orphaned processes; allows 1-click abort of all running scans | Backlog |
| FEAT-012 | 2026-10-27 | UI & Notifications | Persistent notification drawer & OS desktop bread alerts | Notifies hunter when scans complete/stop while working in other windows | Backlog |
| FEAT-013 | 2026-10-30 | Discovered Assets / UI | Dedicated Priority Score & Star Rating Column | Moves priority stars and score values out of the host cell into a dedicated column for easier sorting and visual triage | Implemented |
| FEAT-014 | 2026-10-30 | Subdomain Enumeration | Subdomain Permutation DNS Pipeline (`alterx` piped to `dnsx`) | Eliminates unresolvable permutation garbage; only saves verified live DNS records | Implemented |
| FEAT-015 | 2026-10-30 | Discovered Assets / UI | Subdomain Hierarchical Sitemap & Tree View | Organizes crawled and fuzzed paths (`ffuf`, `katana`, `gau`, `jsluice`) per subdomain in an interactive tree view | Implemented |
| FEAT-016 | 2026-10-30 | Live HTTP Probing / UI | Dedicated Visual Reconnaissance Gallery Tab | Replaces row-by-row modal clicks with a full-screen screenshot card gallery for fast visual triage | Implemented |
| FEAT-017 | 2026-10-30 | Database & Architecture | Discovered Endpoints Schema Separation | Creates dedicated `endpoints` table to prevent crawled paths from overwriting host rows | Implemented |

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
  - Deterministic tool ordering in `ToolsView.tsx` with numbered step badges (`#01`, `#02`, `#03`...) per methodology stage.
  - Category-wise sequential execution ("Run Stage") running all tools in a stage one after another.
  - "Run Full Recon Pipeline" coordinator with optional crawl feedback loop re-probing newly discovered subdomains with `httpx` after crawling & JS mining before fuzzing and vulnerability scanning.
  - Thread-safe scan queue advancement in `App.tsx` responding to WebSocket completion events, with graceful halts on failures or operator stops.

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

#### [FEAT-013] Dedicated Priority Score and Star Rating Column
- **Methodology Stage**: Discovered Assets / Web UI
- **Hunter Justification**: When triaging high-volume asset inventories, hunters prioritize targets based on priority score and star ratings (e.g. dev, staging, admin, login portals). Currently, star ratings are rendered inline within the Host column text, making visual scanning across rows cluttered and preventing direct sorting by priority rank.
- **Architecture**:
  - Table Structure: In `web/src/components/AssetTable.tsx`, introduce a dedicated "Priority" column adjacent to the Host column.
  - Rating Badge: Render priority stars (`★`) alongside numerical score badges with color-coded priority tiers (High `#f0883e`, Medium `#d29922`, Low `#8b949e`).
  - Table Sorting: Enable 1-click column header sorting (`asc` / `desc`) on `asset.priority_score`.
  - Backend Alignment: Ensure `GET /api/assets/target/{targetID}?sort=priority&order=desc` supports direct SQL ordering by `priority_score`.

#### [FEAT-014] Subdomain Permutation DNS Pipeline (`alterx` -> `dnsx`)
- **Methodology Stage**: Subdomain Enumeration & Active Resolution
- **Hunter Justification**: Raw permutation engines (`alterx`, `gotator`) generate thousands of hypothetical naming permutations (e.g. `dev-api`, `admin-test`, `staging-1`). Standalone execution floods asset inventory with unresolvable strings (99.9% noise). Pipelining generated permutations directly into active DNS resolution (`dnsx`) guarantees that only verified, live resolving subdomains are stored.
- **Architecture**:
  - Chained Execution: Modify `internal/tools/alterx.go` to pipe stdout permutations directly into `dnsx -silent -a -resp` or require DNS verification before inserting into `assets`.
  - Storage Integrity: Only record subdomains that returned a valid DNS A or CNAME record.
  - Clean Inventory: Prevents junk wordlist entries from polluting the operator's Discovered Assets view.

#### [FEAT-015] Subdomain Hierarchical Sitemap & Tree View
- **Methodology Stage**: Crawling & Endpoint Discovery / UI Redesign
- **Hunter Justification**: Flat asset tables cannot represent the hierarchical URL directory tree of modern web applications. Hunters need to explore crawled routes (`/api`, `/admin`, `/login`, `/panel`, static assets) organized under their parent subdomain.
- **Architecture**:
  - Interactive Tree Component: In `web/src/components/`, build an expandable sitemap tree component grouping paths by subdomain:
    - `api.target.com`
      - `/v1`
        - `/users` [GET, POST]
        - `/auth` [POST]
      - `/swagger.json` [200 OK]
  - Visual Indicators: Status badges (200, 301, 401, 403), HTTP method chips, and query parameter tags.
  - Quick Filter: Search bar to filter by route path, extension (`.php`, `.json`, `.js`), or status code.

#### [FEAT-016] Dedicated Visual Reconnaissance Gallery Tab
- **Methodology Stage**: Live HTTP Probing / Visual Triage
- **Hunter Justification**: Clicking camera buttons row-by-row in a data table is slow and tedious when triaging 50+ web services. Hunters need a dedicated visual gallery grid where all screenshots are rendered simultaneously for rapid visual triage within seconds.
- **Architecture**:
  - View Toggle: Add view switcher buttons `[ Table View | Visual Gallery | Sitemap Tree ]` to the Discovered Assets header.
  - Visual Card Grid: Render a responsive grid of website thumbnail cards with high-contrast borders:
    - Live thumbnail image with click-to-expand lightbox preview.
    - Status code pill (`200 OK` in green, `403 Forbidden` in yellow, `500` in red).
    - Page Title and detected technologies badges (`React`, `Nginx`, `WordPress`).
    - Quick action buttons: Copy URL, Open in new tab, and Launch Nuclei scan.

#### [FEAT-017] Discovered Endpoints Schema Separation
- **Methodology Stage**: Database & Architecture / Single Responsibility
- **Hunter Justification**: The `assets` table has a `UNIQUE(target_id, host, port)` constraint. When fuzzers (`ffuf`) or crawlers (`katana`, `gau`, `uro`, `jsluice`) discover multiple endpoints on the same host, each newly discovered path overwrites the previous URL on that host row, losing critical endpoints.
- **Architecture**:
  - SQLite Table `endpoints`:
    `CREATE TABLE endpoints (id TEXT PRIMARY KEY, target_id TEXT, host TEXT, url TEXT UNIQUE, path TEXT, method TEXT, status_code INTEGER, content_length INTEGER, title TEXT, source_tool TEXT, created_at DATETIME)`
  - Tool Output Routing: Direct `ffuf`, `katana`, `gau`, `uro`, and `jsluice` discovered routes into `endpoints` rather than `assets`.
  - API & Frontend: Expose `GET /api/endpoints/target/{targetID}` and render in dedicated tab and sitemap tree view.

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
