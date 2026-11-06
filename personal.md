# BHF — Personal Tracking and Operator Log

This document is your personal scratchpad and tracking dashboard for recording bugs, feature ideas, and operational notes while actively using the platform on targets.

---

## 1. Active Bug Backlog

Track defects, crashes, parser anomalies, or UI issues discovered during live execution.

| ID | Date | Assessment Stage | Severity | Summary | Status |
|---|---|---|---|---|---|
| BUG-001 | 2026-10-26 | Core | Low | Terminal autoscroll pauses when clicking inside container | Closed |
| BUG-002 | 2026-10-26 | Vulnerability Assessment | High | Nuclei scan duration takes excessive time; scanner-driven noise | Retired |
| BUG-003 | 2026-10-26 | Discovered Assets / UI | Medium | Discovered Assets pagination shows empty intermediate pages when filtering by Live Web | Resolved |
| BUG-004 | 2026-10-30 | UI / Notifications | Low | Scan stop failure toast message is truncated preventing operator from reading error details | Resolved |
| BUG-005 | 2026-10-30 | Content Discovery | High | Arjun stability probe raises unhandled AttributeError on string response | Resolved |
| BUG-006 | 2026-10-30 | Vulnerability Assessment | High | Nuclei scan fails with missing template dependency | Closed |
| BUG-007 | 2026-10-30 | Tooling / Subdomain Enum | Low | Remove alterx and asnmap tool integrations from framework | Resolved |
| BUG-008 | 2026-10-30 | Live HTTP Probing / Visual | High | Screenshot engine instability (gowitness capture drops/crashes) and Visual Gallery placeholder clutter | Resolved |

*Status Values*: `Open`, `Investigating`, `In Progress`, `Resolved`, `Closed`, `Retired`

### Detailed Bug Descriptions

#### [BUG-002] Nuclei Scan Excessive Execution Time & Tool Deprecation
- **Date**: 2026-10-26
- **Stage**: Vulnerability Assessment
- **Status**: Retired / Deprecated
- **Summary**: Nuclei scans took excessive execution time due to scanning broad default template suites. 
- **Resolution / Architectural Realignment**: Nuclei template scanning has been deprecated and retired from BHF's core architecture. Blind scanner execution is replaced by the deterministic Intelligent Security Assessment engine (`internal/assessment/`) and interactive application intelligence (`internal/appintel/`), which perform targeted, verified checks with reproducible evidence rather than speculative spray-and-pray scanning.

#### [BUG-003] Discovered Assets Pagination and Filter Discrepancy
- **Date**: 2026-10-26
- **Stage**: Discovered Assets / Web UI
- **Resolution**: Migrated filtering to server-side SQL queries (`ListByTarget` in `internal/database/asset_repo.go`) supporting `filter=live`, `2xx`, `3xx`, `errors`, `priority`, and `new`. Updated `AssetHandler.go` to parse `filter`, `sort`, and `order` query parameters.

#### [BUG-004] Scan Stop Failure Truncated Toast Message
- **Date**: 2026-10-30
- **Stage**: UI & Process Controls
- **Resolution**: Replaced single-line CSS `truncate` with `break-words whitespace-pre-wrap max-h-48 overflow-y-auto` and widened `ToastContainer` to `max-w-md`, allowing multi-line error details to be read cleanly.

#### [BUG-005] Arjun Stability Probe Unhandled String Exception
- **Date**: 2026-10-30
- **Stage**: Content Discovery & Parameter Fuzzing
- **Resolution**: Updated `docker/Dockerfile` to install Arjun from git master (`git+https://github.com/s0md3v/Arjun.git`), incorporating upstream fix `if isinstance(response_1, str): return 'skipped'`.

#### [BUG-006] Nuclei Missing Templates Execution Failure
- **Date**: 2026-10-30
- **Stage**: Vulnerability Assessment
- **Status**: Closed / Retired
- **Summary**: Nuclei failed due to missing `/root/nuclei-templates` in runtime environments.
- **Resolution / Architectural Realignment**: Closed due to tool deprecation. External template dependencies are eliminated as part of replacing third-party scanner wrappers with BHF's native, self-contained deterministic assessment checks.

#### [BUG-007] Remove "alterx" and "asnmap" Tool Integrations
- **Date**: 2026-10-30
- **Stage**: Tooling & Subdomain Enumeration
- **Resolution**: Removed `internal/tools/alterx.go` and `internal/tools/asnmap.go`, unregistered tools, and cleaned Dockerfile/tests.

#### [BUG-008] Visual Gallery Screenshot Failures, Placeholders, and gowitness Engine Instability
- **Date**: 2026-10-30
- **Stage**: Live HTTP Probing / Visual Reconnaissance
- **Resolution**: Replaced `gowitness` with ProjectDiscovery's native `httpx -ss` integration in `internal/tools/httpx.go`, unifying live HTTP probing and visual screenshots into a single, high-reliability tool.

---

## 2. Feature & Tool Wishlist

Track capabilities, tool integrations, and workflow improvements that increase assessment quality and research ROI.

| ID | Date | Methodology Stage | Proposed Capability / Tool | Security Assessment ROI / Justification | Status |
|---|---|---|---|---|---|
| FEAT-001 | 2026-10-26 | Live HTTP Probing | TLS certificate SAN extraction (tlsx) | Finds hidden dev/origin hostnames on single IP addresses | Implemented |
| FEAT-002 | 2026-10-26 | Content Discovery | Custom header injection in Ffuf (Auth / Cookie) | Enables authenticated endpoint brute forcing | Implemented |
| FEAT-003 | 2026-10-26 | Crawling & Endpoints | Dedicated Endpoints table & UI tab | Separates crawled web routes from host assets; enables route search | Implemented |
| FEAT-004 | 2026-10-26 | Execution Engine | Real-time streaming ingestion | Pipes stdout lines directly into DB and WS live during execution | Backlog |
| FEAT-005 | 2026-10-26 | Crawling & Endpoints | JavaScript route & secret mining (jsluice) | Uncovers hidden API routes and leaked tokens from crawled JS | Implemented |
| FEAT-006 | 2026-10-26 | HTTP Probing / Vuln | 403/401 bypass engine (nomore403) | Automates header tampering and path mutations | Superseded |
| FEAT-007 | 2026-10-26 | Content Discovery | Parameter discovery (arjun) | Discovers hidden query and body parameters for injection testing | Implemented |
| FEAT-008 | 2026-10-26 | Live HTTP Probing | Visual reconnaissance gallery (httpx -ss) | Automated web screenshots and UI visual triage gallery grid | Implemented |
| FEAT-009 | 2026-10-26 | Progressive Pipeline | Continuous recon delta engine | Historical snapshots and diffing ([NEW], [CHANGED], [REMOVED]) | Backlog |
| FEAT-010 | 2026-10-26 | Workflow Orchestration| Automated multi-stage pipeline chaining | 1-click chained scans across stages with operator approval | Implemented |
| FEAT-011 | 2026-10-27 | Execution Engine | Subprocess group isolation & emergency kill switch | Prevents orphaned processes; allows 1-click abort of all running scans | Backlog |
| FEAT-012 | 2026-10-27 | UI & Notifications | Persistent notification drawer & OS desktop bread alerts | Notifies operator when scans complete/stop while working in other windows | Backlog |
| FEAT-013 | 2026-10-30 | Discovered Assets / UI | Dedicated Priority Score & Star Rating Column | Moves priority stars and score values out of the host cell into a dedicated column | Implemented |
| FEAT-014 | 2026-10-30 | Subdomain Enumeration | Subdomain Permutation DNS Pipeline (`alterx` -> `dnsx`) | Eliminates unresolvable permutation garbage; only saves verified live DNS records | Implemented |
| FEAT-015 | 2026-10-30 | Discovered Assets / UI | Subdomain Hierarchical Sitemap & Tree View | Organizes crawled and fuzzed paths per subdomain in an interactive tree view | Implemented |
| FEAT-016 | 2026-10-30 | Live HTTP Probing / UI | Dedicated Visual Reconnaissance Gallery Tab | Full-screen screenshot card gallery for fast visual triage | Implemented |
| FEAT-017 | 2026-10-30 | Database & Architecture | Discovered Endpoints Schema Separation | Creates dedicated `endpoints` table to prevent crawled paths from overwriting host rows | Implemented |
| FEAT-018 | 2026-11-06 | Interactive App Intel | Headless Browser Exploration Subsystem | Isolated browser session manager via CDP for deep page navigation and DOM state mapping | Implemented |
| FEAT-019 | 2026-11-06 | Interactive App Intel | HTTP Traffic Interception & Categorization Engine | Intercepts HTTP transactions, categorizes flows (auth, admin, upload, API), and maps parameter graph | Implemented |
| FEAT-020 | 2026-11-06 | Intelligent Assessment | Deterministic Security Assessment Engine | Modular verification checks (auth logic, IDOR/BOLA, upload, injection) with reproducible evidence | Implemented |
| FEAT-021 | 2026-11-06 | AI Integration | Hybrid AI Assessment Planner & MCP Tool Suite | Multi-provider AI interface for traffic comprehension, test recommendations, and deterministic gating | Implemented |
| FEAT-022 | 2026-11-06 | Network Security | Network Reconnaissance & Service Discovery | Subnet scanning, host identification, port/service enumeration, OS detection, unified asset model | Implemented |
| FEAT-023 | 2026-11-06 | Network Security | Active Directory & Identity Assessment Module | Kerberos/LDAP enumeration, domain trust mapping, privilege escalation analysis, controlled validation | Implemented |
| FEAT-024 | 2026-11-06 | Interactive App Intel | Traffic Interception & Parameter Graph UI (`TrafficView`) | Visual traffic inspection, category filtering, parameter graph triage, cURL copy | Implemented |
| FEAT-025 | 2026-11-06 | Interactive App Intel | Autonomous Web Exploration & Dynamic DOM Crawler (`AutoExplorer`) | Autonomously visits site, renders JavaScript DOM via Chromium CDP, clicks buttons/forms, intercepts all network traffic into SQLite | Implemented |
| FEAT-026 | 2026-11-06 | Multi-Discipline UI & Workflow | Unified Multi-Discipline UI & Dynamic Cross-Discipline Handoffs | Re-architects UI into Web, Network, AD, Findings, and Console with 1-click cross-discipline routing | Implemented |
| FEAT-027 | 2026-11-06 | UI/UX & Operator Comfort | Eye-Comfort Deep Dark Redesign & Visual De-Cluttering | Replaces harsh borders, rainbow badges, and 50+ button clutter with a calm obsidian palette, unified command bars, and serene tool rows | Implemented |

*Status Values*: `Idea`, `Planned`, `In Progress`, `Implemented`, `Superseded`, `Rejected`

### Detailed Feature Specifications

#### [FEAT-006] 403/401 Forbidden Bypass Engine (`nomore403`)
- **Status**: Superseded
- **Note**: Blind header and path mutation scanning via external `nomore403` binary is superseded by Layer C's deterministic access control and authorization verification checks (`internal/assessment/auth_check.go`).

#### [FEAT-018] Interactive Browser Exploration Subsystem
- **Methodology Layer**: Layer B (Interactive Application Intelligence)
- **Security Assessment Justification**: Static crawling tools miss client-side single page applications (SPAs), state transitions, dynamic form submissions, and authentication handshakes. An isolated, automated browser session provides deep application understanding.
- **Architecture**:
  - Headless browser manager in `internal/appintel/browser.go` utilizing Chrome DevTools Protocol (CDP).
  - Isolated browser sessions per target with configured proxy and network listeners.
  - Automation primitives: page navigation, DOM inspection, form extraction, button clicking, and workflow traversal.
  - Configurable depth, timeout, and interaction limits.

#### [FEAT-019] HTTP Traffic Interception & Categorization Subsystem
- **Methodology Layer**: Layer B (Interactive Application Intelligence)
- **Security Assessment Justification**: Assessment requires understanding what endpoints do rather than just where they are. Structured traffic collection preserves request-response relationships, parameters, and application context.
- **Architecture**:
  - Traffic interceptor in `internal/appintel/interceptor.go` capturing full request/response pairs (method, URL, headers, body, status, timing).
  - Categorization engine in `internal/appintel/categorizer.go` identifying:
    - Authentication (login, logout, session refresh, MFA)
    - Account lifecycle (registration, password reset, account recovery)
    - Data input & search (search queries, filters, user-generated content)
    - File operations (uploads, downloads)
    - Administrative operations (admin panels, settings, user management)
    - API endpoints (REST/GraphQL, JSON payloads)
  - Parameter mapping engine in `internal/appintel/parameter_map.go` tracking user-controlled fields.
  - Sensitive data protection: Automatic redaction of credentials, auth tokens, and PII in logs and UI previews.
  - SQLite table `traffic_records` and `TrafficRepo` in `internal/database/`.

#### [FEAT-020] Deterministic Security Assessment Engine
- **Methodology Layer**: Layer C (Intelligent Security Assessment)
- **Security Assessment Justification**: Replaces speculative scanner alerts with verified, reproducible findings. High reliability over quantity.
- **Architecture**:
  - Modular engine in `internal/assessment/` executing test rules implementing `assessment.Rule`.
  - Core check modules:
    - `auth_check.go`: Authentication parameters, session handling, rate limiting controls.
    - `idor_check.go`: Object identifier mapping, cross-account access boundary verification.
    - `input_check.go`: Deterministic injection and reflection checks.
    - `upload_check.go`: File type validation, extension handling, access controls.
  - Verification Gate: Every finding requires verified proof of impact, confidence grading (`CONFIRMED`, `HIGH`, `MEDIUM`, `INFORMATIONAL`), and reproducible `curl` commands and raw request/response pairs.

#### [FEAT-021] Hybrid AI Assessment Planner & MCP Tool Suite
- **Methodology Layer**: AI Integration & Layer C
- **Security Assessment Justification**: AI agents excel at understanding complex multi-step application logic and recommending attack surfaces, but should never be a single point of failure.
- **Architecture**:
  - Multi-provider AI interface in `internal/ai/` supporting Gemini, Claude, and local LLMs.
  - Analysis modules: traffic comprehension, parameter classification, test recommendation.
  - Interactive MCP tool suite: exposes tools for agents to inspect categorized traffic, request browser actions, and trigger deterministic assessment tests.
  - Deterministic Gating: All AI hypotheses must be validated through deterministic checks before a finding is recorded.

#### [FEAT-022] Network Reconnaissance & Service Discovery Subsystem
- **Methodology Layer**: Network Security Expansion
- **Security Assessment Justification**: Penetration testing covers infrastructure alongside web targets. Web and network assessment must share the same asset inventory and evidence models.
- **Architecture**:
  - Network discovery in `internal/network/discovery.go` for local subnet and host identification.
  - Port & service prober in `internal/network/service_probe.go` with service fingerprinting, banner grabbing, and OS detection.
  - Automatic handoff: Discovered HTTP/HTTPS services automatically register into the Web Recon and Application Intelligence layers, linking web findings to the parent host asset.

#### [FEAT-023] Active Directory & Identity Assessment Module
- **Methodology Layer**: Network Security Expansion
- **Security Assessment Justification**: Active Directory is the central identity backbone in enterprise networks. Mapping AD relationships, Kerberos vulnerabilities, and misconfigurations provides immense pentest value.
- **Architecture**:
  - AD mapper in `internal/network/ad_mapper.go` operating via LDAP/Kerberos protocols within explicitly authorized assessment scopes.
  - Auditing: Domain trusts, unconstrained delegation, kerberoasting exposure, privileged group memberships.
  - Controlled post-exploitation validation with strict safety guards and zero destructive actions.

#### [FEAT-025] Autonomous Web Exploration & Dynamic DOM Crawler (`AutoExplorer`)
- **Methodology Layer**: Layer B (Interactive Application Intelligence)
- **Security Assessment Justification**: Static web crawlers miss single-page applications (SPAs), state transitions, dynamic client-side forms, and interactive tabs. An autonomous crawler with CDP and DOM evaluation exercises the full attack surface automatically without manual browser interaction.
- **Architecture**:
  - Dual-mode crawler in `internal/appintel/auto_explorer.go`.
  - Headless Chromium CDP mode binds to an internal recording proxy, navigates pages, executes DOM interactions (clicks buttons, fills inputs with benign probes, submits forms, scrolls), and records 100% of network traffic into `traffic_records`.
  - Pure-Go HTTP spider fallback mode parses HTML links and forms when Chromium is absent.
  - REST API endpoint (`POST /api/appintel/target/{id}/auto-explore`) and MCP tool (`auto_explore_target`).

#### [FEAT-026] Unified Multi-Discipline UI Architecture & Dynamic Cross-Discipline Handoffs
- **Methodology Layer**: UI/UX & Methodology Orchestration
- **Security Assessment Justification**: As BHF expands into web, network, and Active Directory security, the user interface must reflect the exact domain of assessment rather than cluttering tools together. Discovered network assets must dynamically seed web recon and AD audits.
- **Architecture**:
  - Re-architected primary navigation into five top-level disciplines in `Navbar.tsx`: **Web Security**, **Network Security**, **Active Directory**, **Findings & Evidence**, and **Live Terminal**.
  - Unified `WebSecurityView.tsx` consolidating Methodology & Tools, Discovered Assets & Visual Gallery, and Interactive App Intel & Traffic Explorer under sub-tabs.
  - Dedicated `NetworkSecurityView.tsx` for CIDR/IP discovery, port profiling, service fingerprinting, and misconfiguration alerting.
  - Dedicated `ActiveDirectoryView.tsx` for Domain Controller detection, LDAP/Kerberos transport security, and policy compliance.
  - Dynamic 1-click discipline handoffs: web ports discovered during network scans route directly to Web Security with adjusted scope; AD ports route directly to Active Directory security audits.

---

## 3. Operational Checklist: Testing the MCP Server

Use this testing checklist to verify Model Context Protocol (MCP) server integration:

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
  *Expected*: Returns JSON list of supported tools (`list_targets`, `list_assets`, `list_endpoints`, `list_findings`, `get_target_stats`, `list_tools`) with input schemas.

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

- [ ] **Test 5: Call `list_assets` and `list_endpoints`**
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

## 4. Live Assessment Notes & Target Observations

Operator scratchpad for noting quirks, WAF behaviors, rate limits, or interesting application behaviors during authorized assessments:

- **Assessment Notes**:
  - Keep rate limit for targets under 50 req/sec to avoid triggering CDN/WAF rate-limiting.
  - Review `401` and `403` endpoints for HTTP method manipulation (`GET`, `POST`, `HEAD`, `OPTIONS`) and path normalization behaviors.
  - Check for API version discrepancies (`/api/v1` vs `/api/v2`) where older endpoints lack modern authorization checks.
