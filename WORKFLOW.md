# BHF — Bug Hunting & Security Assessment Platform: Development Workflow and Operational Runbook

This document defines the standard engineering lifecycle for adding features, integrating assessment modules, managing dependencies, debugging defects, and tracking progress from an offensive security research perspective.

---

## 1. Security Assessment Methodology: The Progressive Assessment Pipeline

BHF is designed around an extensible, progressive assessment funnel. The framework continuously evolves, allowing new assessment capabilities, application intelligence stages, and specialized modules to be integrated cleanly:

```
[Perimeter Discovery]     -> Discover in-scope perimeter, subdomains & resolve DNS
       |
       v
[Network & Port Mapping]  -> Discover network hosts, non-standard ports & open services
       |
       v
[HTTP/S Probing]          -> Fingerprint live web targets, status codes, tech stack & priority
       |
       v
[Crawling & Routes]       -> Extract dynamic routes, API endpoints, JS secrets & parameters
       |
       v
[Content Fuzzing]         -> Surgical fuzzing for sensitive files, backups & unlinked routes
       |
       v
[Interactive App Intel]   -> Isolated browser navigation, traffic interception & categorization
       |
       v
[Intelligent Assessment]  -> Deterministic verification + AI analysis with reproducible evidence
       |
       v
[Network & AD Expansion]  -> Service audit, Active Directory mapping & privilege validation
```

### Stage 1: Subdomain and OSINT Enumeration
- **Primary Tools**: `subfinder`, `assetfinder`, `theharvester`, `dnsx`
- **Methodology**: Expand the perimeter as broadly as possible without triggering active alerts. Scrape certificate transparency logs, search engines, and passive threat feeds. Immediately resolve records via `dnsx` to drop dead hosts and wildcard DNS false positives.
- **Critical Invariant**: Strip wildcards and enforce `tools.IsInScope()` so out-of-scope third parties never contaminate the target inventory.

### Stage 2: Port and Network Scanning
- **Primary Tools**: `naabu`
- **Methodology**: Production web servers on standard ports 80 and 443 are often fronted by CDNs and WAFs. High-yield services (unauthenticated dev consoles, staging APIs, Prometheus/Actuator metrics, internal dashboards) frequently hide on non-standard ports (e.g., `8080`, `8443`, `8000`, `9000`, `3000`, `5000`, `9200`).
- **Traffic Safety**: Scan only resolved, live hostnames. Never flood targets with uncontrolled SYN bursts that trigger network-level rate limits.

### Stage 3: Live HTTP/S Probing and Fingerprinting
- **Primary Tools**: `httpx`, `tlsx`, `wafw00f`
- **Methodology**: Determine which open ports actually serve HTTP/HTTPS traffic. Inspect HTTP response status codes, page titles, server headers, and technology stacks. Extract SANs to uncover origin servers.
- **Priority Scoring (ROI Optimization)**:
  - Automatically flag high-value assets with elevated priority scores:
    - Status codes `401` and `403`: High-priority authentication and access control candidates.
    - Sensitive keywords in hostname: `admin`, `api`, `internal`, `dev`, `staging`, `vpn`, `auth`, `portal`.
    - Sensitive technologies: Spring Boot, Laravel, Jenkins, Swagger, Kubernetes, Actuator.
    - Non-standard ports: `8080`, `8443`, `8000`, `9000`, `3000`.

### Stage 4: Crawling and URL Discovery
- **Primary Tools**: `katana`, `gau`, `uro`, `jsluice`
- **Methodology**: Spider dynamic applications to extract endpoints, hidden JavaScript routes, and query parameters. Mine historical archives (Wayback Machine, Common Crawl, AlienVault) for forgotten endpoints and deprecated API routes. Parse JavaScript bundles using tree-sitter AST to uncover hidden endpoints and credentials.
- **Scope Guard**: Constrain crawlers to in-scope domains (`-fs dn`) to avoid spidering external CDNs, social media widgets, or third-party auth providers. Discovered routes are routed into the `endpoints` table.

### Stage 5: Content Discovery and Parameter Fuzzing
- **Primary Tools**: `ffuf`, `arjun`
- **Methodology**: Perform surgical directory and parameter discovery against live web targets. Seek high-impact exposures: configuration dumps, backup archives, unlinked portals, and hidden query/body parameters (`id=`, `user_id=`, `role=`, `file=`, `redirect=`).
- **Traffic Discipline**: Never use reckless defaults sending thousands of req/sec. Use calibrated rate limits (`-rate 80`), sensible wordlists, and status filtering.

### Stage 6: Interactive Application Intelligence
- **Subsystem**: `internal/appintel/` (Browser controller, traffic proxy, transaction categorizer)
- **Methodology**: Launch an isolated, headless browser session to navigate target applications. Observe real user flows, state transitions, authentication forms, and client-side API calls. Intercept HTTP transactions and categorize them into authentication, registration, user-generated content, file upload, search, and admin flows. Map all user-controlled parameters into a structured, searchable traffic graph.
- **Privacy & Safety**: Redact and protect sensitive session tokens and personal information (PII).

### Stage 7: Intelligent Security Assessment & Deterministic Verification
- **Subsystem**: `internal/assessment/` + `internal/ai/` (Hybrid AI planner + deterministic engine)
- **Methodology**: Move beyond speculative scanner alerts. AI agents analyze captured traffic and recommend high-probability security checks based on real application functionality. Deterministic testing engines execute targeted tests (auth bypass, IDOR/BOLA, unsafe file uploads, parameter injection, access control boundary violations).
- **Evidence Requirement**: Findings are created ONLY upon deterministic verification. Every finding includes an explicit confidence level, full request/response evidence, and an executable reproduction command.

### Stage 8: Network Infrastructure & Active Directory Assessment
- **Subsystem**: `internal/network/`
- **Methodology**: Enumerate network subnets, discover internal hosts and running services, map network topology, audit Active Directory domain relationships, and detect protocol misconfigurations within explicitly authorized scopes.

---

## 2. Feature Addition Lifecycle

Every code modification or feature addition must follow this 7-step sequence:

```
[1. Security Value & Scope Assessment] 
       |
       v
[2. Architecture & SRP Placement]
       |
       v
[3. Test Definition]
       |
       v
[4. Clean Implementation]
       |
       v
[5. Rigorous Verification (go test & npm build)]
       |
       v
[6. Changelog Documentation]
       |
       v
[7. Atomic Conventional Git Commit]
```

### Step 1: Security Value, Safety & Scope Assessment
Before writing any code, answer these mandatory questions:
1. **Which methodology pipeline stage or layer does this serve?** Features that do not yield high-signal intelligence, provide verified evidence, or speed up analysis are rejected.
2. **Will this trigger excessive network traffic or WAF bans?** Active network requests must include rate limiting (`-rl`), concurrency bounds (`-c`), and timeout controls.
3. **Is scope strictly maintained?** Ensure that no third-party assets can leak into the database.
4. **Is evidence reproducible?** Ensure any findings generated are backed by exact request/response data and verified impact.

### Step 2: Architecture, Single Responsibility, and DRY Placement
Determine exactly which files need to be created or modified according to Single Responsibility Principle (SRP) and Don't Repeat Yourself (DRY):
- New recon tool? Add `internal/tools/<tool_name>.go` implementing `tools.Tool`. Reuse existing wrappers and universal scope guards (`tools.IsInScope()`).
- New application intelligence module? Place under `internal/appintel/`.
- New deterministic security check? Place under `internal/assessment/` implementing `assessment.Rule`.
- New network discovery module? Place under `internal/network/`.
- New database entity? Add `internal/models/<entity>.go`, `internal/database/<entity>_repo.go`, update canonical schema in `internal/database/schema.go`, and add `web/src/types/<entity>.ts`.
- New UI panel? Create a dedicated view in `web/src/views/` and granular components in `web/src/components/`. All network communications must route through dedicated services in `web/src/services/`.

### Step 3: Test Definition
Define tests for the new functionality before or alongside code implementation:
- If adding a tool, test its output parser in `internal/tools/tools_test.go` with sample strings.
- If adding an assessment check, test its verification logic in `internal/assessment/*_test.go` using mock HTTP response fixtures.
- If modifying database logic, add a test in `internal/database/<entity>_repo_test.go`.

### Step 4: Clean Implementation and Single Source of Truth (SSOT)
- Write idiomatic, clean Go code using typed structs, explicit error checking (`if err != nil`), and clean interfaces.
- Enforce DRY: Never duplicate business rules, domain normalization, priority calculations, or validation logic across files. Refer to the canonical authoritative source.
- Avoid premature abstraction: Follow the Rule of Three; only extract shared helper utilities when identical domain knowledge is repeated across 3 or more locations.
- Write modern TypeScript React code using functional components with explicit types and dedicated service clients.
- Adhere strictly to the GitHub Dark UI theme (`#0d1117`) with zero neon colors, zero fluff text, and zero emojis.

### Step 5: Rigorous Verification
Run all automated test suites to ensure zero regressions:
```bash
go test -v ./...
go test -race ./...
cd web && npm run build && cd ..
```

### Step 6: Progress Tracking
Record the change in `CHANGELOG.md` under the `[Unreleased]` section following the standard categories (`Added`, `Changed`, `Fixed`, `Removed`, `Security`).

### Step 7: Atomic Conventional Git Commit
Commit the changes using a descriptive Conventional Commit message:
```bash
git add <specific-files>
git commit -m "feat(assessment): implement deterministic idor boundary verification check"
```

---

## 3. Python Tool Isolation Runbook

To add or update a Python-based security tool (e.g., `theHarvester`, `wafw00f`):

1. **Never use global pip**: Installing packages into the host OS or global container environment is strictly forbidden.
2. **Use uv tool isolation**:
   Inside the Dockerfile, install tools into their own virtual environment using `uv`:
   ```dockerfile
   RUN uv tool install --python 3.14 git+https://github.com/laramies/theHarvester.git
   ```
3. **Subprocess Execution in Go**:
   Configure the tool's `BinaryName()` in `internal/tools/<tool>.go` to check for the installed executable:
   ```go
   func (t *Tool) BinaryName() string {
       if _, err := exec.LookPath("tool-name"); err == nil {
           return "tool-name"
       }
       return "uv"
   }
   ```
4. **Local Development Virtual Environment**:
   If developing Python helper scripts locally, create a project-local virtual environment:
   ```bash
   uv venv .venv
   source .venv/bin/activate
   ```
   Ensure `.venv` is ignored by `.gitignore`.

---

## 4. Methodical Investigation Runbook (Zero Guesswork)

When an unexpected defect occurs:

### Step 1: Capture the Exact Output and Command
- Review the terminal logs for the scan session via the Console tab or SQLite:
  ```bash
  sqlite3 data/bhf.db "SELECT stream, line FROM logs WHERE scan_id = 'TARGET_SCAN_ID';"
  ```
- Identify the exact command line arguments and whether the binary returned a non-zero exit code.

### Step 2: Validate the Subprocess and Binary Environment
- Test whether the binary is in `$PATH`:
  ```bash
  which <tool-name>
  ```
- Run the tool directly with the same arguments against a test target to observe raw stdout and stderr behavior.

### Step 3: Inspect Database Records and Constraints
- Check SQLite for table schema mismatches or foreign key violations:
  ```bash
  sqlite3 data/bhf.db ".schema scans"
  sqlite3 data/bhf.db "SELECT * FROM scans WHERE id = 'TARGET_SCAN_ID';"
  ```

### Step 4: Write an Isolating Unit Test
- Create a test case reproducing the failing output or invalid input inside the corresponding `_test.go` file.
- Verify that the test fails as expected.

### Step 5: Implement the Fix and Confirm Resolution
- Fix the root cause in the source code.
- Run `go test ./...` and verify the reproducing test now passes.
- Test in the running application to confirm end-to-end resolution.

---

## 5. Starter Prompt for New Conversations

When opening a new session or thread with the AI assistant, copy and paste the following prompt. It establishes the security assessment philosophy and enforces all engineering standards:

```text
You are working on BHF, an extensible security assessment and penetration testing automation platform located at /home/diamond/Desktop/Github/bhf.

Before proposing or making any changes, you must read:
1. ARCHITECTURE.md (system design, models, directory structure, MCP)
2. RULES.md (engineering rules, UI standards, Single Responsibility Principle, zero emojis, zero guesswork)
3. TESTING.md (test suite architecture and table-driven test patterns)
4. WORKFLOW.md (security methodology, pipeline stages, tool isolation, debugging runbook)
5. CHANGELOG.md (current progress and recent modifications)
6. personal.md (operator tracking, active bugs, and feature wishlist)

Core Philosophy:
Reliability over quantity. Evidence over assumptions. Verification over speculation.
We are building a professional security engineering platform, not a collection of scripts that blindly throw payloads at targets.

Target Architecture & Major Layers:
- Layer A: Reconnaissance & Discovery (Web reconnaissance and network discovery normalized into shared models).
- Layer B: Interactive Application Intelligence (Isolated browser session, traffic interception, categorization into auth, admin, uploads, search, APIs, parameter mapping, and state transitions).
- Layer C: Intelligent Security Assessment (Hybrid AI planning + deterministic verification engines, reproducible request/response evidence, explicit confidence levels).
- AI Integration: Hybrid architecture (deterministic engines for repeatability and validation; AI agents for comprehension, traffic analysis, and test planning; multi-provider support).
- Network Security Expansion: Local network discovery, host/service enumeration, Active Directory assessment, privilege mapping, controlled validation, sharing unified asset/evidence store.

Core Engineering Commandments:
- Follow the Single Responsibility Principle: each file has one single job.
- Use Conventional Commits naming methodology (feat, fix, refactor, test, docs, chore).
- UI must remain clean, minimal, developer-grade GitHub Dark (#0d1117). Strictly NO AI glow, NO neon tropes, NO emojis.
- Every assessment check must directly serve the security assessment cause with safe defaults and calibrated rate limits.
- Never report unverified scanner alerts as findings; every finding must include reproducible evidence and verified impact.
- Never guess causes of bugs; investigate methodically using logs and reproduction.
- Ensure all Go tests pass via 'go test ./...' before completing your task.
- Track all changes in CHANGELOG.md.

Current Task:
<DESCRIBE YOUR TASK HERE>
```
