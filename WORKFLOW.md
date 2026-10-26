# BHF (Bug Hunting Framework) - Development Workflow and Operational Runbook

This document defines the standard engineering lifecycle for adding features, integrating security tools, managing dependencies, debugging defects, and tracking progress from an offensive bug hunter's perspective.

---

## 1. The Hunter's Methodology: 6 Offensive Reconnaissance Sections

BHF is designed around the progressive recon funnel used by professional bug bounty hunters. Every feature, scanner, and data model must map directly to one of these 6 sections:

```
[Section 1: Subdomain & OSINT] -> Discover in-scope perimeter & resolve DNS
       |
       v
[Section 2: Port & Network]   -> Uncover non-standard ports & exposed services
       |
       v
[Section 3: HTTP/S Probing]    -> Fingerprint live web, status codes, tech stack, priority score
       |
       v
[Section 4: Crawling & URLs]   -> Extract dynamic routes, API endpoints, sensitive query params
       |
       v
[Section 5: Content Discovery] -> Targeted fuzzing for .git, .env, backups, admin portals
       |
       v
[Section 6: Vulnerability Scan]-> High-confidence template execution & verified cURL reproduction
```

### Section 1: Subdomain and OSINT Enumeration
- **Primary Tools**: `subfinder`, `assetfinder`, `theharvester`, `dnsx`
- **Hunter Mindset**: Expand the perimeter as broadly as possible without triggering active alerts. Scrape certificate transparency logs, search engines, and passive threat feeds. Immediately resolve records via `dnsx` to drop dead hosts and wildcard DNS false positives.
- **Critical Invariant**: Strip wildcards and enforce `tools.IsInScope()` so out-of-scope third parties never contaminate the target inventory.

### Section 2: Port and Network Scanning
- **Primary Tools**: `naabu`
- **Hunter Mindset**: Production web servers on ports 80 and 443 are heavily hardened and protected by enterprise WAFs. High-yield bugs (unauthenticated admin panels, dev consoles, staging APIs, Prometheus/Actuator metrics) commonly hide on non-standard ports (e.g., `8080`, `8443`, `8000`, `9000`, `3000`, `5000`, `9200`).
- **Traffic Safety**: Scan only resolved, live hostnames. Never flood targets with uncontrolled SYN bursts that trigger network-level rate limits.

### Section 3: Live HTTP/S Probing and Fingerprinting
- **Primary Tools**: `httpx`
- **Hunter Mindset**: Determine which open ports actually serve HTTP/HTTPS traffic. Inspect HTTP response status codes, page titles, redirects, and technology stacks.
- **Priority Scoring (ROI Optimization)**:
  - Automatically flag high-value assets with elevated priority scores:
    - Status codes `401` and `403`: High-priority authentication bypass and forbidden-bypass candidates.
    - Sensitive keywords in hostname: `admin`, `api`, `internal`, `dev`, `staging`, `vpn`, `auth`, `portal`.
    - Sensitive technologies: Spring Boot, Laravel, Jenkins, Swagger, Kubernetes, Actuator, Apache Solr.
    - Non-standard ports: `8080`, `8443`, `8000`, `9000`, `3000`.

### Section 4: Crawling and URL Discovery
- **Primary Tools**: `katana`, `gau`
- **Hunter Mindset**: Spider dynamic applications to extract endpoints, hidden JavaScript routes, and sensitive query parameters (`id=`, `url=`, `redirect=`, `token=`, `file=`). Mine historical archives (Wayback Machine, Common Crawl, AlienVault) for forgotten endpoints and deprecated API routes.
- **Scope Guard**: Constrain crawlers to in-scope domains (`-fs dn`) to avoid spidering external CDNs, social media widgets, or auth providers.

### Section 5: Content Discovery and Fuzzing
- **Primary Tools**: `ffuf`
- **Hunter Mindset**: Perform surgical directory and file brute-forcing against live web targets. Seek high-impact exposures: `.git/config`, `.env`, backup files (`.bak`, `.zip`), configuration dumps, and unlinked admin panels.
- **Traffic Discipline**: Never use brute-force defaults that send 5,000 req/sec and trigger immediate WAF IP bans. Use calibrated rate limits (`-rate 80`), sensible wordlists, and status filtering (`-mc 200,204,301,302,307,401,403`).

### Section 6: Vulnerability Assessment and Verification
- **Primary Tools**: `nuclei`
- **Hunter Mindset**: Scan discovered endpoints with targeted, high-signal vulnerability templates (CVEs, misconfigurations, exposed panels, default credentials).
- **Reproduction Invariant**: Every finding must include a verified, executable reproduction command (`curl`) and clear severity. Never report low-severity junk, informational banner grabs, or unverified false positives.

---

## 2. Feature Addition Lifecycle

Every code modification or feature addition must follow this 7-step sequence:

```
[1. Recon Value Assessment] 
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

### Step 1: Reconnaissance Value and Traffic Safety Assessment
Before writing any code, answer these three mandatory questions:
1. **Which of the 6 hunting sections does this serve?** Features that do not yield high-signal intelligence or speed up analysis are rejected.
2. **Will this trigger excessive network traffic or WAF bans?** Any active network requests must include rate limiting (`-rl`), concurrency bounds (`-c`), and timeout controls.
3. **Is scope strictly maintained?** Ensure that no third-party assets can leak into the database.

### Step 2: Architecture and Single Responsibility Placement
Determine exactly which files need to be created or modified according to Single Responsibility Principle:
- New security tool? Add exactly one file in `internal/tools/<tool_name>.go` implementing `tools.Tool`.
- New database entity? Add `internal/models/<entity>.go`, `internal/database/<entity>_repo.go`, update `internal/database/schema.go`, and add `web/src/types/<entity>.ts`.
- New UI panel? Create a dedicated view in `web/src/views/` and granular components in `web/src/components/`.

### Step 3: Test Definition
Define tests for the new functionality before or alongside code implementation:
- If adding a tool, prepare sample output lines and test its `Run()` parser in `internal/tools/tools_test.go`.
- If modifying database logic, add a test in `internal/database/<entity>_repo_test.go`.

### Step 4: Clean Implementation
- Write idiomatic, clean Go code using typed structs, explicit error checking (`if err != nil`), and clean interfaces.
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
git commit -m "feat(tools): add katana crawler integration with domain scoping"
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

When a bug occurs (e.g., scan fails unexpectedly, output is not parsed, database error):

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
- Run the tool directly with the same arguments against a test domain to observe raw stdout and stderr behavior.

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

When opening a new session or thread with the AI assistant, copy and paste the following prompt. It instructs the assistant to adopt the hunter's perspective, follow the 6 offensive reconnaissance sections, and adhere strictly to all engineering rules:

```text
You are working on BHF (Bug Hunting Framework) located at /home/diamond/Desktop/Github/bhf.

Before proposing or making any changes, you must read:
1. ARCHITECTURE.md (system design, models, directory structure, MCP)
2. RULES.md (engineering rules, UI standards, Single Responsibility Principle, zero emojis, zero guesswork)
3. TESTING.md (test suite architecture and table-driven test patterns)
4. WORKFLOW.md (hunter methodology, 6 recon sections, tool isolation, debugging runbook)
5. CHANGELOG.md (current progress and recent modifications)
6. personal.md (operator tracking, active bugs, and feature wishlist)

The Hunter's Perspective and 6 Offensive Sections:
Always approach tasks with a bug bounty hunter's mindset (prioritizing high ROI over noise, stealth, and actionable reproduction). Structure your thinking around the 6 core recon sections:
1. Subdomain & OSINT Enumeration (subfinder, assetfinder, theharvester, dnsx)
2. Port & Network Scanning (naabu)
3. Live HTTP/S Probing & Fingerprinting (httpx, priority scoring)
4. Crawling & URL Discovery (katana, gau)
5. Content Discovery & Fuzzing (ffuf)
6. Vulnerability Assessment & Verification (nuclei)

Core Engineering Commandments:
- Follow the Single Responsibility Principle: each file has one single job.
- Use Conventional Commits naming methodology (feat, fix, refactor, test, docs, chore).
- UI must remain clean, minimal, developer-grade GitHub Dark (#0d1117). Strictly NO AI glow, NO neon tropes, NO emojis.
- Every feature must directly serve the bug hunting cause without generating excessive or reckless traffic.
- Never guess causes of bugs; investigate methodically using logs and reproduction.
- Ensure all Go tests pass via 'go test ./...' before completing your task.
- Track all changes in CHANGELOG.md.

Current Task:
<DESCRIBE YOUR TASK HERE>
```
