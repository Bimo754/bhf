# BHF (Bug Hunting Framework) - System Architecture

This document provides a comprehensive technical blueprint of the BHF architecture, detailing component interactions, data pipelines, domain models, and design decisions.

---

## 1. System Overview

BHF is a fast, lightweight, and modular reconnaissance framework engineered for professional bug bounty hunters and penetration testers. It replaces bloated, multi-container architectures (like Ars0n Framework) with a high-speed, single-binary distribution powered by a Go backend, modern embedded SQLite in WAL mode, and a modular React/TypeScript frontend.

### Core Architectural Characteristics

- **Single-Container or Standalone Binary**: Operates either inside a lightweight ~350 MB Docker image or as a static, single executable with the frontend embedded via Go `embed.FS`.
- **Pure Go SQLite (Zero CGO)**: Uses `modernc.org/sqlite` with Write-Ahead Logging (`WAL`) mode, foreign keys, and 5000ms busy timeout for high-concurrency read/write operations without CGO compilation dependencies.
- **Real-Time Log Streaming**: Native WebSockets deliver stdout and stderr lines instantaneously to the browser console while simultaneously persisting them into SQLite.
- **Strict Scope Integrity**: Enforces root domain and subdomain scoping guards (`tools.IsInScope()`) on all discovered hostnames and endpoints before storage.
- **Model Context Protocol (MCP)**: Native JSON-RPC 2.0 interface supporting both stdio (`--mcp`) and HTTP (`POST /api/mcp`) for seamless integration with AI coding assistants and autonomous agents.

---

## 2. High-Level Architecture Diagram

```
+---------------------------------------------------------------------------------+
|                               Web Browser / Client                              |
|   React + TypeScript + Vite | GitHub Dark Theme | WebSocket Event Consumer     |
+---------------------------------------+-----------------------------------------+
                                        |
                   HTTP API / REST      |      WebSockets (/api/ws/scans/{id}/logs)
                                        v
+---------------------------------------------------------------------------------+
|                                 Go HTTP Server                                  |
|   Chi Router | CORS | Middleware (RequestID, RealIP, Logger, Recoverer)        |
+-------------------+-------------------+-------------------+---------------------+
                    |                   |                   |
                    v                   v                   v
           +-----------------+  +-----------------+  +------------------+
           | Target & Scans  |  | Discovered      |  | Model Context    |
           | Handlers        |  | Assets/Findings |  | Protocol (MCP)   |
           +--------+--------+  +--------+--------+  +--------+---------+
                    |                    |                    |
                    v                    v                    v
+---------------------------------------------------------------------------------+
|                            Execution Engine & Hub                               |
|   - Dispatcher: Manages active scans and cancelation contexts                   |
|   - Runner: Executes subprocesses, streams lines through bufio.Scanner           |
|   - Hub: Thread-safe WebSocket connection registry and broadcaster              |
+---------------------------------------+-----------------------------------------+
                                        |
               Spawns Isolated Subprocesses & Parses Output
                                        v
+---------------------------------------------------------------------------------+
|                             Security Tools Pipeline                             |
|   Subfinder | Assetfinder | theHarvester (uv) | DNSx | Naabu | Httpx | Katana  |
|   Gau | Ffuf | Nuclei                                                           |
+---------------------------------------+-----------------------------------------+
                                        |
                     Clean & Validated Findings (IsInScope)
                                        v
+---------------------------------------------------------------------------------+
|                      Persistence Layer (SQLite in WAL Mode)                     |
|   Repositories: TargetRepo | ScanRepo | AssetRepo | FindingRepo | LogRepo       |
|   Tables: targets | scans | assets | findings | logs                            |
+---------------------------------------------------------------------------------+
```

---

## 3. Directory Layout and File Responsibilities

```
bhf/
├── cmd/
│   └── bhf-server/
│       └── main.go               # Application entrypoint, CLI flags, dependency wiring
├── internal/
│   ├── api/                      # HTTP presentation layer (Chi router and handlers)
│   │   ├── asset_handler.go      # Discovered assets pagination and filtering
│   │   ├── finding_handler.go    # Vulnerability findings and severity stats
│   │   ├── response.go           # Standard JSON response envelope and error formatters
│   │   ├── router.go             # Route registration and middleware configuration
│   │   ├── scan_handler.go       # Scan lifecycle management (start, stop, logs)
│   │   ├── target_handler.go     # Target program scope creation, retrieval, deletion
│   │   ├── tool_handler.go       # Tool registry status catalog
│   │   └── ws_handler.go         # WebSocket upgrade and subscription handling
│   ├── config/                   # Configuration management
│   │   ├── config.go             # Config struct definition
│   │   └── loader.go             # Environment variable parsing and directory preparation
│   ├── database/                 # SQLite storage layer (pure Go, WAL mode)
│   │   ├── asset_repo.go         # Asset persistence, priority scoring, statistics
│   │   ├── connection.go         # Database handle initialization and PRAGMA tuning
│   │   ├── finding_repo.go       # Finding insertion, listing, severity metrics
│   │   ├── log_repo.go           # Scan log persistence and retrieval
│   │   ├── scan_repo.go          # Scan history, status updates, arguments
│   │   ├── schema.go             # DDL migrations, tables, indexes
│   │   └── target_repo.go        # Target CRUD operations
│   ├── engine/                   # Process execution and streaming engine
│   │   ├── dispatcher.go         # Thread-safe scan registration and cancellation
│   │   ├── hub.go                # WebSocket broadcast manager
│   │   ├── runner.go             # Low-level subprocess execution and line streaming
│   │   └── stream.go             # Real-time message contracts
│   ├── mcp/                      # Model Context Protocol server (AI integration)
│   │   └── server.go             # JSON-RPC 2.0 stdio and HTTP handler
│   ├── models/                   # Core domain entities
│   │   ├── asset.go              # Discovered host, URL, port, tech entity
│   │   ├── finding.go            # Vulnerability finding entity
│   │   ├── log.go                # Log entry entity (stdout/stderr)
│   │   ├── scan.go               # Scan execution session entity
│   │   └── target.go             # Bug bounty target scope entity
│   └── tools/                    # Security tool integrations
│       ├── alterx.go             # Alterx subdomain permutation generator
│       ├── asnmap.go             # Asnmap ASN and CIDR mapping tool
│       ├── assetfinder.go        # Assetfinder runner and parser
│       ├── dnsx.go               # DNSx DNS resolution and A/CNAME record parser
│       ├── ffuf.go               # Ffuf web fuzzer and wordlist manager
│       ├── gau.go                # GetAllURLs archive extractor
│       ├── httpx.go              # Httpx live probing and technology parser
│       ├── interface.go          # Tool interface, Category, IsInScope, CleanDomain
│       ├── katana.go             # Katana crawler and endpoint extractor
│       ├── naabu.go              # Naabu port scanner integration
│       ├── nuclei.go             # Nuclei vulnerability scanner and JSONL parser
│       ├── registry.go           # Thread-safe registry for installed tools
│       ├── subfinder.go          # Subfinder passive subdomain enumerator
│       ├── subzy.go              # Subzy subdomain takeover vulnerability checker
│       ├── theharvester.go       # theHarvester OSINT gatherer (uv isolation)
│       ├── tlsx.go               # Tlsx TLS SAN and certificate intelligence tool
│       └── wafw00f.go            # Wafw00f WAF detection and fingerprinting tool (uv isolation)
│   ├── web/                      # Frontend single-page application
│   │   ├── embed.go              # Go embed.FS declaration for dist/
│   │   ├── src/
│   │   │   ├── components/       # Granular, reusable React UI components
│   │   │   ├── services/         # API and WebSocket client services
│   │   │   ├── types/            # Domain TypeScript interface declarations
│   │   │   ├── views/            # Top-level view containers (Methodology, Assets, Findings, Console)
│   │   │   ├── App.tsx           # Root component, state coordination, toasts
│   │   │   ├── index.css         # Tailwind base styles and selection overrides
│   │   │   └── main.tsx          # React DOM initialization
│   │   ├── package.json          # Web dependencies
│   │   ├── tailwind.config.js    # Design tokens and custom colors
│   │   └── vite.config.ts        # Vite build configuration
│   ├── docker/
│   │   └── Dockerfile            # Multi-stage production container build
│   ├── docker-compose.yml        # Container deployment configuration
│   └── Makefile                  # Build automation targets
│   ```
│   
│   ---
│   
│   ## 4. Execution Pipeline and Methodology Stages
│   
│   Reconnaissance in BHF is organized into standard offensive security stages:
│   
│   1. **Subdomain and OSINT Enumeration**:
│      - `subfinder`: Passive DNS and multi-source public query.
│      - `assetfinder`: Public archive and asset discovery.
│      - `theharvester`: Search engine and certificate log scraping inside an isolated virtual environment (`uv`).
│      - `alterx`: Fast rule-based subdomain wordlist permutation generator.
│      - `asnmap`: Network range, ASN, and CIDR block mapping.
│      - `dnsx`: Multi-resolver DNS resolution, wildcard filtering, and IP extraction.
│   2. **Port and Network Scanning**:
│      - `naabu`: Fast SYN/Connect port scanning across discovered hostnames to identify exposed non-standard services.
│   3. **HTTP Probing**:
│      - `httpx`: Probes discovered hosts on HTTP/HTTPS ports, extracts HTTP status codes, page titles, and web technologies.
│      - `tlsx`: Extracts SAN hostnames from certificates and captures JARM fingerprints.
│      - `wafw00f`: Identifies protective WAF vendor technologies to tune rate limits.
│   4. **Crawling and URL Discovery**:
│      - `katana`: Spider dynamic endpoints, JS files, and API routes with `-fs dn` scope protection.
│      - `gau`: Fetches known endpoints from Wayback Machine, AlienVault, and Common Crawl.
│   5. **Content Discovery and Fuzzing**:
│      - `ffuf`: Directory and endpoint brute-forcing with calibrated rate limits against discovered live web targets.
│   6. **Vulnerability Assessment**:
│      - `subzy`: Subdomain takeover vulnerability checker auditing dangling CNAME records.
│      - `nuclei`: Fast, template-based scanning for exposures, misconfigurations, and known CVEs. Output is streamed live and parsed into structured findings with reproduction `curl` commands.

---

## 5. Persistence and Data Models

The SQLite database enforces referential integrity through foreign keys with cascading deletions:

```
+-----------------------------------+
|              targets              |
|-----------------------------------|
| id          TEXT PRIMARY KEY      |
| name        TEXT NOT NULL         |
| domain      TEXT NOT NULL UNIQUE  |
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
+--------+------------------+-------+
         | 1                | 1
         |                  |
         | *                | *
+--------+--------+  +------+---------+  +-------------------+
|     assets      |  |    findings    |  |       logs        |
|-----------------|  |----------------|  |-------------------|
| id              |  | id             |  | id (INTEGER PK)   |
| target_id (FK)  |  | target_id (FK) |  | scan_id (FK)      |
| scan_id (FK)    |  | scan_id (FK)   |  | stream (stdout/err|
| host            |  | tool_name      |  | line (TEXT)       |
| url             |  | severity       |  | timestamp         |
| ip              |  | name           |  +-------------------+
| port            |  | matched_at     |
| status_code     |  | description    |
| title           |  | extracted_res  |
| technologies    |  | curl_command   |
| priority_score  |  | created_at     |
| created_at      |  +----------------+
+-----------------+
```

### Asset Prioritization Engine

Assets inserted or updated via `AssetRepo.Upsert` are automatically assigned a dynamic `priority_score`:
- Baseline: `0`
- High-value technology detected (e.g., Spring Boot, Jenkins, Laravel, Kubernetes, Swagger, Actuator): `+25`
- Sensitive keywords in hostname (e.g., `admin`, `api`, `internal`, `dev`, `staging`, `vpn`, `auth`, `portal`): `+20`
- Non-standard ports (e.g., `8080`, `8443`, `8000`, `9000`, `3000`): `+15`
- Status codes indicating access restrictions or interesting endpoints (`401`, `403`): `+10`

Assets with a score `>= 30` are flagged as **Priority Targets** in the operator interface.

---

## 6. Model Context Protocol (MCP) Integration

BHF implements the Model Context Protocol to expose reconnaissance data directly to AI agents:

- **Protocol Version**: JSON-RPC 2.0 (`protocolVersion: 2024-11-05`)
- **Transport Modes**:
  1. **Standard I/O**: `bhf-server --mcp` reads from `stdin` and writes JSON-RPC responses to `stdout`.
  2. **HTTP API**: `POST /api/mcp` accepts JSON-RPC request payloads and returns JSON-RPC responses.
- **Available MCP Tools**:
  - `list_targets`: Retrieve all registered bug bounty target domains.
  - `list_assets`: Query discovered hosts, live URLs, ports, and technologies for a target.
  - `list_findings`: Query identified vulnerabilities, severities, and reproduction commands.
  - `get_target_stats`: Get aggregated asset counts and severity breakdowns.
  - `list_tools`: Retrieve all supported reconnaissance tools and binary status.
