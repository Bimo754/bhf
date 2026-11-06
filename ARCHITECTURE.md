# BHF — Bug Hunting & Security Assessment Platform: System Architecture

This document provides the canonical architectural blueprint of the BHF platform, detailing system subsystems, inter-component communication, data models, evidence lifecycles, and engineering design decisions.

---

## 1. System Overview & Vision

BHF is an extensible, evidence-driven security assessment and penetration testing automation platform. Engineered for authorized security researchers, penetration testers, and application security engineers, BHF unifies the entire assessment lifecycle into a cohesive, high-reliability environment.

BHF moves beyond traditional reconnaissance runners that blindly execute external command-line tools and report raw, unverified output. Instead, BHF understands its targets: it discovers perimeter attack surfaces, interacts intelligently with web applications, intercepts and categorizes network traffic, maps parameters and state transitions, selects targeted security tests, deterministically validates potential vulnerabilities, and preserves reproducible evidence.

### Core Architectural Characteristics

- **Reliability Over Quantity, Evidence Over Assumptions, Verification Over Speculation**: BHF rejects scanner-driven noise and speculative alerts. Every reported finding must possess an explicit confidence level, verified reproducibility (e.g., exact HTTP request/response evidence or executable reproduction command), and documented security impact.
- **Single-Container or Standalone Go Distribution**: Operates as a static, single binary with embedded React/TypeScript frontend via Go `embed.FS`, or inside a lightweight multi-stage Docker container without multi-gigabyte bloat or microservice sprawl.
- **Pure Go SQLite (WAL Mode, Zero CGO Runtime)**: Uses `modernc.org/sqlite` with Write-Ahead Logging (`WAL`), cascading foreign keys, and tuned busy timeouts for high-concurrency read/write operations with zero external database dependencies.
- **Hybrid AI & Deterministic Automation**: Couples deterministic engines for repeatable tasks, normalization, and verification with AI agents for application comprehension, traffic analysis, and test prioritization. AI hypotheses are strictly gated by deterministic validation.
- **Model Context Protocol (MCP)**: Native JSON-RPC 2.0 interface supporting both standard I/O (`--mcp`) and HTTP (`POST /api/mcp`), enabling autonomous AI agents to interact with the target inventory, browser exploration sessions, and assessment engines.
- **Scope & Data Integrity**: Universal scoping guards enforce root domain, subdomain, CIDR, and IP range constraints before any asset, endpoint, or finding is recorded.

---

## 2. High-Level Architecture Diagram

```
+---------------------------------------------------------------------------------------------------+
|                                       Operator & Agent Layer                                      |
|   React + TypeScript UI (GitHub Dark #0d1117)  |  WebSocket Event Consumer  |  External AI via MCP|
+-------------------------------------------------+-------------------------------------------------+
                                                  |
                         HTTP REST API            |            WebSockets & MCP (JSON-RPC 2.0)
                                                  v
+---------------------------------------------------------------------------------------------------+
|                                          Go Server Engine                                         |
|   Chi Router | CORS | Auth & Scope Middleware | SSE / WebSocket Hub | MCP JSON-RPC Server         |
+-------------------------------------------------+-------------------------------------------------+
                                                  |
         +----------------------------------------+----------------------------------------+
         |                                        |                                        |
         v                                        v                                        v
+------------------------+      +----------------------------------+      +-------------------------+
|  A. Recon & Discovery  |      |   B. Interactive App Intel       |      |   C. Security Assessment|
|  - Web Reconnaissance  |      |   - Headless Browser Controller  |      |   - Hybrid AI Planner   |
|  - Network & Service   |      |   - Traffic Interceptor & Proxy  |      |   - Deterministic Engine|
|    Discovery           |      |   - Route & Parameter Categorizer|      |   - Evidence Validator  |
+-----------+------------+      +-----------------+----------------+      +------------+------------+
            |                                     |                                    |
            +-------------------------------------+------------------------------------+
                                                  |
                                                  v
+---------------------------------------------------------------------------------------------------+
|                                  D. Network Security Subsystem                                    |
|   Host Inventory | Port & Service Enumerator | Active Directory & Identity Mapper | Post-Exploit   |
+-------------------------------------------------+-------------------------------------------------+
                                                  |
                                                  v
+---------------------------------------------------------------------------------------------------+
|                              E. Unified Execution & Persistence Layer                             |
|   Dispatcher (Contexts & Jobs) | Process Runner (Group Isolation) | SQLite WAL Persistence       |
|   Repositories: TargetRepo | HostRepo | ServiceRepo | AssetRepo | EndpointRepo | TrafficRepo     |
|                 FindingRepo | ScanRepo | LogRepo                                                  |
+---------------------------------------------------------------------------------------------------+
```

---

## 3. Directory Layout and File Responsibilities

The codebase enforces strict Single Responsibility Principle (SRP): one file, one defined responsibility.

```
bhf/
├── cmd/
│   └── bhf-server/
│       └── main.go                     # Entrypoint, CLI flags, dependency injection, graceful shutdown
├── internal/
│   ├── api/                            # HTTP presentation layer (Chi router, handlers, middleware)
│   │   ├── asset_handler.go            # Discovered host assets, priority filtering, stats
│   │   ├── endpoint_handler.go         # Discovered HTTP routes, sitemap endpoints, path queries
│   │   ├── finding_handler.go          # Vulnerability findings, severity metrics, evidence export
│   │   ├── response.go                 # Standard JSON envelopes and unified error formatters
│   │   ├── router.go                   # Route registration, middleware, static file serving
│   │   ├── scan_handler.go             # Scan lifecycle (launch, cancel, skip tool, status)
│   │   ├── target_handler.go           # Target scope management (domains, CIDRs, IPs)
│   │   ├── tool_handler.go             # Tool registry status catalog and capabilities
│   │   └── ws_handler.go               # WebSocket upgrade and real-time subscription hub
│   ├── appintel/                       # Interactive Application Intelligence (Layer B)
│   │   ├── auto_explorer.go            # Autonomous dynamic application explorer & DOM interaction crawler
│   │   ├── browser.go                  # Headless browser manager (CDP / Chromium automation)
│   │   ├── categorizer.go              # Traffic classifier (auth, admin, api, upload, UGC)
│   │   ├── interceptor.go              # HTTP traffic interception and transaction capture
│   │   ├── parameter_map.go            # Input parameter graph and state transition tracker
│   │   └── session.go                  # Cookie, token, and session context manager (PII protected)
│   ├── assessment/                     # Intelligent Security Assessment Engine (Layer C)
│   │   ├── auth_check.go               # Authentication, session handling, and bypass testing
│   │   ├── engine.go                   # Assessment coordinator, test runner, rate limiter
│   │   ├── evidence.go                 # Reproducible evidence builder (cURL, raw HTTP transaction)
│   │   ├── idor_check.go               # Object identifier mapping and authorization boundary tests
│   │   ├── input_check.go              # Deterministic injection and input validation verifier
│   │   ├── rule.go                     # Assessment test interface, confidence scoring, criteria
│   │   └── upload_check.go             # File upload validation, content-type, and storage tests
│   ├── ai/                             # AI Orchestration and Agent Integrations
│   │   ├── client.go                   # Multi-provider LLM client interface (Gemini, Claude, OpenAI)
│   │   ├── planner.go                  # Attack surface prioritization and test recommendation
│   │   └── traffic_analyzer.go         # Endpoint comprehension and anomaly detection
│   ├── network/                        # Network Security Subsystem
│   │   ├── ad_mapper.go                # Active Directory, LDAP, and Kerberos enumeration
│   │   ├── discovery.go                # Local subnet discovery, ARP/ICMP scanning, host identification
│   │   ├── misconfig.go                # Network service misconfiguration and privilege checks
│   │   ├── service_probe.go            # Service banner grabber, version detection, OS fingerprinting
│   │   └── topology.go                 # Network exposure and infrastructure relationship mapping
│   ├── config/                         # Configuration and environment management
│   │   ├── config.go                   # System configuration struct definition
│   │   └── loader.go                   # Environment parsing, directory bootstrapping, validation
│   ├── database/                       # Persistence layer (Pure Go SQLite in WAL mode)
│   │   ├── asset_repo.go               # Host-level asset persistence, scoring, filtering
│   │   ├── connection.go               # Database handle initialization, PRAGMA tuning
│   │   ├── endpoint_repo.go            # Discovered URL routes, query parameters, status codes
│   │   ├── finding_repo.go             # Finding persistence, evidence verification, metrics
│   │   ├── log_repo.go                 # Real-time stdout/stderr scan log lines
│   │   ├── scan_repo.go                # Scan execution state, arguments, timestamps
│   │   ├── schema.go                   # Canonical DDL migrations, tables, foreign keys, indexes
│   │   └── target_repo.go              # Target program scope CRUD
│   ├── engine/                         # Process execution, cancellation, and streaming engine
│   │   ├── dispatcher.go               # Thread-safe scan registration and cancellation contexts
│   │   ├── hub.go                      # WebSocket broadcast manager and client registry
│   │   ├── runner.go                   # Subprocess group execution, streaming, and termination
│   │   └── stream.go                   # Real-time log message contracts
│   ├── mcp/                            # Model Context Protocol server (AI integration)
│   │   └── server.go                   # JSON-RPC 2.0 stdio and HTTP handler exposing BHF tools
│   ├── models/                         # Domain entities
│   │   ├── asset.go                    # Discovered host, IP, port, technologies entity
│   │   ├── endpoint.go                 # Discovered URL route, HTTP method, status entity
│   │   ├── finding.go                  # Vulnerability finding with reproducible evidence
│   │   ├── log.go                      # Execution log line entity
│   │   ├── scan.go                     # Tool execution session entity
│   │   ├── target.go                   # Bug bounty and assessment target scope entity
│   │   └── traffic.go                  # Intercepted HTTP transaction entity
│   └── tools/                          # External reconnaissance CLI adapters
│       ├── arjun.go                    # Arjun parameter discovery tool (uv isolated)
│       ├── assetfinder.go              # Assetfinder runner and output parser
│       ├── corsy.go                    # Corsy CORS misconfiguration scanner (uv isolated)
│       ├── dnsx.go                     # DNSx DNS resolver, wildcard elimination, A/CNAME parser
│       ├── ffuf.go                     # Ffuf web fuzzer and wordlist manager
│       ├── gau.go                      # GetAllURLs historical archive extractor
│       ├── httpx.go                    # Httpx live probing, technology parser, screenshot engine
│       ├── interface.go                # Tool interface, Category, IsInScope, CleanDomain
│       ├── jsluice.go                  # Jsluice tree-sitter AST JavaScript route and secret parser
│       ├── katana.go                   # Katana crawler and endpoint extractor
│       ├── naabu.go                    # Naabu port scanner integration
│       ├── registry.go                 # Thread-safe registry for installed tools
│       ├── s3scanner.go                # S3Scanner cloud storage bucket reconnaissance tool
│       ├── subfinder.go                # Subfinder passive subdomain enumerator
│       ├── subzy.go                    # Subzy subdomain takeover vulnerability checker
│       ├── theharvester.go             # theHarvester OSINT gatherer (uv isolated)
│       ├── tlsx.go                     # Tlsx TLS SAN and certificate intelligence tool
│       ├── uro.go                      # Uro URL decluttering and deduplication filter (uv isolated)
│       └── wafw00f.go                  # Wafw00f WAF detection and fingerprinting tool (uv isolated)
├── web/                                # Frontend single-page application (React + Vite)
│   ├── embed.go                        # Go embed.FS declaration for dist/ assets
│   ├── src/
│   │   ├── components/                 # Reusable UI components (AssetTable, SitemapTree, etc.)
│   │   ├── services/                   # Typed API and WebSocket client services
│   │   ├── types/                      # Domain TypeScript interfaces matching backend models
│   │   ├── views/                      # Discipline views (WebSecurity, NetworkSecurity, ActiveDirectory, Findings, Terminal)
│   │   ├── App.tsx                     # Root component, layout, active scan synchronization
│   │   ├── index.css                   # Tailwind base styles and high-contrast selections
│   │   └── main.tsx                    # React DOM initialization
│   ├── package.json                    # Web dependencies
│   ├── tailwind.config.js              # GitHub Dark design tokens
│   └── vite.config.ts                  # Vite build configuration
├── docker/
│   └── Dockerfile                      # Multi-stage production container build
├── docker-compose.yml                  # Container deployment configuration
└── Makefile                            # Build automation targets
```

---

## 4. Target Architecture: The Three Core Layers

BHF is structured into three primary assessment layers, augmented by a hybrid AI orchestration system and an infrastructure network security subsystem.

```
+----------------------------------------------------------------------------------------------------+
|  Layer A: Reconnaissance & Discovery                                                               |
|  - Web: Subdomain enumeration, DNS wildcard filtering, live HTTP probing, tech stack, crawling    |
|  - Network: Subnet discovery, host discovery, port enumeration, service fingerprinting, OS ID     |
+----------------------------------------------------------------------------------------------------+
                                                  |
                               Discovered Perimeter & Attack Surface
                                                  v
+----------------------------------------------------------------------------------------------------+
|  Layer B: Interactive Application Intelligence                                                    |
|  - Isolated browser exploration, DOM interaction, page workflow navigation                        |
|  - Traffic interception: structured request/response capture, header/body parsing                 |
|  - Categorization: auth, registration, reset, search, upload, UGC, admin, API routes               |
|  - Parameter mapping: state transitions, user inputs, sensitive data protection                   |
+----------------------------------------------------------------------------------------------------+
                                                  |
                             Structured & Categorized Traffic Graph
                                                  v
+----------------------------------------------------------------------------------------------------+
|  Layer C: Intelligent Security Assessment                                                          |
|  - Hybrid AI Planner: traffic analysis, endpoint comprehension, targeted test recommendations      |
|  - Deterministic Engines: auth checks, injection validation, IDOR mapping, upload evaluation       |
|  - Verification & Evidence: confidence grading, reproducible cURL / raw HTTP, verified impact      |
+----------------------------------------------------------------------------------------------------+
```

### Layer A: Reconnaissance and Discovery

Maps the external and internal attack surface with high speed and strict scoping.

1. **Web Reconnaissance**:
   - **Passive & Active Subdomain Enumeration**: `subfinder`, `assetfinder`, `theharvester`.
   - **DNS Resolution & Wildcard Elimination**: `dnsx` resolves records, extracts A/CNAME data, and eliminates wildcards.
   - **Live HTTP/S Probing & Tech Fingerprinting**: `httpx` identifies active web servers, HTTP status codes, server headers, technologies, and captures headless visual screenshots.
   - **TLS & Origin Intelligence**: `tlsx` extracts Subject Alternative Names (SANs) and JARM fingerprints to reveal backend origin hosts.
   - **Crawling & Historical URL Extraction**: `katana` (crawler with `-fs dn` boundary guards), `gau` (Wayback, Common Crawl, AlienVault), `uro` (URL decluttering), and `jsluice` (tree-sitter AST extraction of routes and secrets from JavaScript bundles).
   - **Content Discovery & Parameter Fuzzing**: `ffuf` (calibrated directory brute-forcing) and `arjun` (hidden query and body parameter discovery).

2. **Network Reconnaissance**:
   - **Local Subnet & Host Discovery**: Identifies live hosts across CIDR blocks and local subnets.
   - **Port & Service Enumeration**: `naabu` and service probers scan TCP/UDP ports without flooding targets.
   - **Service Fingerprinting & OS Detection**: Banner grabbing, service protocol negotiation, and OS heuristics.
   - **Topology Mapping**: Relates IP addresses, subnets, gateways, and infrastructure roles.

Discovery outputs are normalized into shared `hosts`, `services`, `assets`, and `endpoints` database models.

### Layer B: Interactive Application Intelligence

Reconnaissance endpoints indicate *where* an application lives, but not *how it works*. Layer B observes real application behavior through deep traffic analysis and browser interaction.

1. **Autonomous & Interactive Workflow**:
   - Operator selects target and triggers autonomous exploration (`AutoExplorer`).
   - **Autonomous Dynamic Crawler (`AutoExplorer`)**: Visits pages, renders modern JavaScript DOM via Chromium CDP, clicks interactive buttons/tabs, fills forms with safe probe inputs, scrolls dynamic content, and records all background network traffic (XHR, Fetch, Form posts).
   - **Manual Browser Sessions**: Operator can also launch isolated Chromium sessions or route external tools through the built-in interceptor proxy.
   - Intercepts and records all HTTP requests and responses into `traffic_records` automatically without requiring manual proxy interaction.
   - Organizes captured traffic into a searchable, structured traffic graph.

2. **Traffic Categorization**:
   The engine automatically inspects and categorizes transactions:
   - **Authentication**: Login, logout, session refresh, MFA prompts.
   - **Account Lifecycle**: Registration, email verification, password reset, account recovery.
   - **Data Input & Search**: Search queries, filter parameters, user-generated content (UGC).
   - **File Operations**: Multipart file uploads, document downloads, avatar processing.
   - **Administrative**: Admin panels, user management, configuration endpoints, metric exports.
   - **API Operations**: REST/GraphQL operations, JSON payloads, structured error responses.
   - **Session & Identity**: Cookies, JWTs, Bearer tokens, API keys, CSRF tokens.
   - **User-Controlled Parameters**: Query parameters, body parameters, custom headers.

3. **Data Integrity & Protection**:
   - Captures request method, full URL, headers, body, response status, response headers, response body, and timing.
   - Preserves request-response relationships and parent-child interaction sequences.
   - Sensitive credentials, session tokens, and personal data (PII) are masked or encrypted at rest.

### Layer C: Intelligent Security Assessment

Replaces blind scanner execution with evidence-driven assessment. AI agents and deterministic engines collaborate:

```
[Categorized Traffic & Endpoints]
               |
               v
[AI Comprehension & Test Recommendation]  <-- AI Agent analyzes traffic context
               |
               v
[Deterministic Test Execution]           <-- Deterministic engine executes specific test
               |
               v
[Verification Gate]                      <-- Validates behavior against expected criteria
               |
        +------+------+
        |             |
   [Verified]    [Speculative / Unverified]
        |             |
        v             v
[Evidence Capture]  [Logged as Observation Only]
- Raw Request
- Raw Response
- Reproduction cURL
- Confidence Level
- Verified Impact
        |
        v
[Structured Finding]
```

1. **Deterministic Testing Modules**:
   - **Authentication & Session Logic**: Evaluates authentication parameter exposure, session expiration, token handling, and rate limiting controls.
   - **Authorization & IDOR / BOLA**: Maps object identifiers across different user sessions; verifies authorization boundaries between user contexts.
   - **Input Validation & Injection**: Tests user-controlled parameters with controlled, verifiable payloads (SQL injection, reflection, template injection).
   - **File Upload Security**: Assesses file type validation, MIME type spoofing, storage path isolation, and executable execution restrictions.
   - **API Business Logic**: Audits state transitions, method tampering, and sensitive data exposure.

2. **The Verification Gate**:
   - A finding is NEVER created merely because a payload was sent, a status code changed, or an error occurred.
   - Every candidate finding must pass through a verification rule that validates specific security impact.
   - Every finding records:
     - **Confidence Level**: `CONFIRMED` (deterministic proof of exploitability), `HIGH` (strong evidence with minor environmental dependency), `MEDIUM` (suspicious behavior requiring manual verification), or `INFORMATIONAL` (security posture observation).
     - **Reproducible Evidence**: Exact HTTP request and response pair, plus an executable `curl` command.
     - **Clear Security Impact**: Explanation of what an adversary achieves.

---

## 5. AI Integration & Hybrid Automation

BHF implements a hybrid architecture where AI agents and deterministic code reinforce each other:

| Component | Deterministic Engines | AI Agents |
|---|---|---|
| **Role** | Execution, normalization, validation, safety | Comprehension, reasoning, prioritization |
| **Responsibilities** | - Execute repeatable security tests<br>- Normalize network and web data<br>- Enforce rate limits and scope guards<br>- Validate test responses against pass/fail criteria<br>- Generate reproducible evidence | - Understand application functionality<br>- Analyze complex HTTP traffic<br>- Map relationships between endpoints<br>- Classify parameters and user roles<br>- Recommend high-yield security tests<br>- Interpret ambiguous application behavior |
| **Fail-Safe** | Hardcoded boundaries, rate limits, timeouts | AI output is treated as hypothesis until verified |

### Tool-Based Agent Interaction

AI agents interact with BHF subsystems through explicitly defined MCP tools and internal APIs:
- `browser_navigate`: Request the browser to navigate to a target URL.
- `browser_interact`: Fill a form, submit input, or trigger an application event.
- `inspect_traffic`: Query captured HTTP transactions by category, method, or parameter.
- `recommend_assessment`: Propose security tests against identified endpoints.
- `execute_verification`: Run a deterministic verification check against a proposed endpoint.

### Multi-Provider Architecture

The AI subsystem defines a clean adapter interface (`ai.Client`) supporting multiple LLM backends (Gemini, Claude, local models via Ollama/vLLM) without coupling the platform to any single vendor.

---

## 6. Network Security Subsystem

Network security shares the same core asset inventory, evidence store, and reporting infrastructure as web security:

```
[Network Discovery] -> Subnet Scan -> Discovered Hosts -> Open Ports -> Running Services
                                                                              |
                                        +-------------------------------------+--------------------+
                                        |                                                          |
                                        v                                                          v
                                [HTTP/S Web Service]                                    [Infrastructure / AD]
                                        |                                                          |
                                        v                                                          v
                               [Web Assessment Layer]                                     [Network Assessment]
                               - Subdomains & URLs                                        - Active Directory / LDAP
                               - Browser Exploration                                      - Service Misconfiguration
                               - App Security Testing                                     - Privilege Escalation
                                        |                                                          |
                                        +-------------------------------------+--------------------+
                                                                              |
                                                                              v
                                                          [Unified Asset & Finding Store]
```

### Key Responsibilities

1. **Host & Service Discovery**: Enumerates IP ranges, resolves hostnames, identifies listening ports, and fingerprints protocols (SSH, SMB, RDP, Database, HTTP).
2. **Web Attack Surface Handoff**: When open HTTP/HTTPS services are discovered on a host, they are automatically registered into the Web Reconnaissance and Application Intelligence pipeline, associating web findings back to the parent network asset.
3. **Active Directory & Identity Assessment**: Audits Kerberos authentication, LDAP configurations, domain trust relationships, and excessive privileges within authorized boundaries.
4. **Controlled Validation**: Assesses default credentials, anonymous access, and protocol misconfigurations with non-destructive verification tests.

---

## 7. Persistence and Shared Data Models

The SQLite database operates in WAL mode with strict foreign key referential integrity and cascading deletes:

```
+-----------------------------------+
|              targets              |
|-----------------------------------|
| id          TEXT PRIMARY KEY      |
| name        TEXT NOT NULL         |
| domain      TEXT NOT NULL UNIQUE  |
| scope_type  TEXT NOT NULL         | (domain, cidr, ip)
| notes       TEXT                  |
| created_at  DATETIME              |
| updated_at  DATETIME              |
+-----------------+-----------------+
                  | 1
                  |
                  | *
+-----------------+-----------------+
|               scans               |
|-----------------------------------|
| id          TEXT PRIMARY KEY      |
| target_id   TEXT (FK -> targets)  |
| tool_name   TEXT NOT NULL         |
| status      TEXT NOT NULL         |
| command     TEXT NOT NULL         |
| arguments   TEXT (JSON)           |
| started_at  DATETIME              |
| completed_at DATETIME             |
| error_msg   TEXT                  |
+---+-------------+-------------+---+
    | 1           | 1           | 1
    |             |             |
    | *           | *           | *
+---+-----+  +----+----+  +-----+---+   +--------------------+
|  assets |  |endpoints|  |  traffic|   |      findings      |
|---------|  |---------|  |---------|   |--------------------|
| id (PK) |  | id (PK) |  | id (PK) |   | id (PK)            |
| target_id  | target_id  |target_id|   | target_id (FK)     |
| scan_id |  | scan_id |  | scan_id |   | scan_id (FK)       |
| host    |  | host    |  | method  |   | tool_name          |
| url     |  | url (UQ)|  | url     |   | severity           |
| ip      |  | path    |  | req_hdrs|   | confidence_level   | (CONFIRMED, HIGH, etc)
| port    |  | method  |  | req_body|   | name               |
| status  |  | status  |  | res_code|   | matched_at         |
| title   |  | length  |  | res_hdrs|   | description        |
| techs   |  | title   |  | res_body|   | curl_command       |
| priority|  | source  |  | category|   | evidence_request   |
| created |  | created |  | auth_ctx|   | evidence_response  |
+---------+  +---------+  | created |   | impact             |
                          +---------+   | created_at         |
                                        +--------------------+
```

### Evidence Lifecycle

Every security observation progresses through a defined lifecycle:

1. **Surface Observation**: Discovered by Layer A (host, open port, URL) or Layer B (intercepted HTTP request).
2. **Classification**: Categorized by functionality (e.g., authentication endpoint, parameter `user_id`).
3. **Hypothesis & Test Plan**: Assessment rule or AI agent proposes a targeted verification test.
4. **Controlled Execution**: Test is executed with strict rate limiting and timeout controls.
5. **Deterministic Verification**: Response is evaluated against strict verification criteria (e.g., did an unauthorized user retrieve another user's profile data?).
6. **Evidence Compilation**: Full request, response, headers, timing, and executable `curl` command are captured.
7. **Finding Recording**: Persisted to SQLite with assigned severity, confidence level, and actionable remediation notes.

---

## 8. Model Context Protocol (MCP) Integration

BHF exposes a JSON-RPC 2.0 interface (`protocolVersion: 2024-11-05`) for external AI agents:

- **Transport Modes**:
  1. **Standard I/O (`stdio`)**: `bhf-server --mcp` reads JSON-RPC requests from `stdin` and writes to `stdout`.
  2. **HTTP API**: `POST /api/mcp` accepts JSON-RPC request payloads and returns JSON-RPC responses.
- **Available MCP Tools**:
  - `list_targets`: Query registered assessment target scopes.
  - `list_assets`: Query discovered hosts, live URLs, ports, and technologies.
  - `list_endpoints`: Query discovered HTTP paths, methods, and status codes.
  - `list_findings`: Query validated vulnerability findings with evidence and reproduction commands.
  - `get_target_stats`: Retrieve asset metrics, severity breakdowns, and assessment progress.
  - `list_tools`: Retrieve supported reconnaissance and assessment tools with binary status.
  - `query_traffic`: Query captured HTTP transactions by category, method, or parameter.
  - `execute_assessment`: Launch a targeted assessment test with deterministic verification.

---

## 9. Core Engineering Principles: SRP and DRY

BHF strictly adheres to established software engineering disciplines:

### Single Responsibility Principle (SRP)
- **One File, One Responsibility**: Backend repositories, tool adapters, HTTP handlers, frontend components, and domain interfaces are strictly isolated in dedicated files.
- **Explicit Layer Separation**: Presentation handlers never perform database queries directly; subprocess runners never parse tool output; tool parsers never execute database writes; UI components never make direct `fetch()` calls.

### Don't Repeat Yourself (DRY) and Single Source of Truth (SSOT)
- **Scope Verification**: `tools.IsInScope()` and `tools.CleanDomain()` reside exclusively in `internal/tools/interface.go` as the single authoritative scope boundary.
- **Asset Prioritization**: `CalculatePriorityScore()` in `internal/database/asset_repo.go` is the single source of truth for priority scoring heuristics.
- **Database Schema**: `internal/database/schema.go` is the authoritative source for database tables and constraints. Go models and TypeScript types mirror this schema.
- **Pragmatic DRY (The Rule of Three)**: Avoid premature abstractions. Common helpers are extracted only when identical business logic appears across three or more call sites.
