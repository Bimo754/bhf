# BHF (Bug Hunting Framework) - Engineering and Operating Rules

This document establishes the binding engineering principles, architectural constraints, operator preferences, and quality standards for the BHF project. Every contributor, session, and modification must strictly follow these rules without exception.

---

## 1. System and Environment Integrity

- **Zero Host Package Pollution**: Never execute `pacman`, `apt`, `brew`, or install security tools or system libraries directly on the host operating system. Never install global Python `pip` packages.
- **Isolated Execution**: All tools, recon processes, and services must run inside the single, multi-stage Docker container (`docker compose up -d --build`, approximately 350 MB footprint) or inside isolated project-local virtual environments (`uv` or `python -m venv`).
- **Anti-Bloat Architecture**: Never recreate Ars0n's architecture. Avoid multi-gigabyte container proliferation, microservice sprawl, or unnecessary background daemons. Maintain a unified, lightweight, and high-speed execution profile.

---

## 2. Strict Modularity and Single Responsibility Principle (SRP)

- **One File = One Responsibility**: Enforced strictly across backend, frontend, tools, and configurations. It is prohibited to combine disparate functionalities into a single file or create monolithic classes.
- **Backend Directory Boundaries**:
  - `internal/models/`: One domain entity per file (`target.go`, `scan.go`, `asset.go`, `finding.go`, `log.go`).
  - `internal/database/`: One repository struct per domain entity (`target_repo.go`, `scan_repo.go`, `asset_repo.go`, `finding_repo.go`, `log_repo.go`). Connection and schema migration belong in dedicated files (`connection.go`, `schema.go`).
  - `internal/engine/`: Dedicated components for process management (`runner.go`), execution dispatching (`dispatcher.go`), WebSocket distribution (`hub.go`), and message contracts (`stream.go`).
  - `internal/tools/`: Exactly one file per security tool integration (`subfinder.go`, `httpx.go`, `katana.go`, `ffuf.go`, etc.). Tool interfaces and common helper functions reside in `interface.go`, registry in `registry.go`.
  - `internal/api/`: One HTTP handler struct per resource domain (`target_handler.go`, `scan_handler.go`, `asset_handler.go`, `finding_handler.go`, `tool_handler.go`, `ws_handler.go`). Response envelopes reside in `response.go`, routing in `router.go`.
  - `internal/mcp/`: Dedicated Model Context Protocol server exposing assets and findings to external AI agents (`server.go`).
- **Frontend Directory Boundaries**:
  - `web/src/types/`: One TypeScript interface/type file per domain entity (`target.ts`, `scan.ts`, `asset.ts`, `finding.ts`, `log.ts`, `tool.ts`).
  - `web/src/services/`: Dedicated service client per domain (`api.ts`, `targetService.ts`, `scanService.ts`, `assetService.ts`, `findingService.ts`, `toolService.ts`, `wsService.ts`). Never make direct `fetch` calls from inside UI components.
  - `web/src/components/`: Granular, reusable UI components (`ToolCard.tsx`, `Terminal.tsx`, `AssetTable.tsx`, `Toast.tsx`, `ScopeDropdown.tsx`, `Button.tsx`, `Badge.tsx`, `RunConfigModal.tsx`, `CreateTargetModal.tsx`, `DeleteTargetModal.tsx`).
  - `web/src/views/`: Dedicated page views that compose components (`ToolsView.tsx`, `AssetsView.tsx`, `FindingsView.tsx`, `TerminalView.tsx`).
- **No Monolithic Frontends**: Never bundle HTML, CSS, and JavaScript into a tangled blob. Use clean Tailwind classes, separate reusable components, and distinct service layers.

---

## 3. UI/UX Aesthetic and Design Standards

- **Zero AI Tropes**: Strictly NO neon cyan, purple, or magenta glows, gradients, rounded glassy bubbles, or gimmicky animations that make the application look AI-generated.
- **Developer-Grade Dark Theme**: Strict GitHub Dark / Linear dark palette:
  - Canvas / Background: `#0d1117`
  - Cards, panels, elevated surfaces: `#161b22`
  - Borders and dividers: `#30363d`
  - Active and hover states: `#21262d`
  - Body text: `#e6edf3`
  - Muted / secondary text: `#8b949e`
- **High-Contrast Selection**: Text selection (`::selection`) must always render high contrast (`#1f6feb` background with `#ffffff` text). Low-contrast selections where text becomes illegible are strictly prohibited.
- **No Native HTML Form Chrome**: Never use unstyled browser `<select>` dropdowns or generic inputs. Use custom dark-themed components (e.g., `ScopeDropdown`).
- **No Fluff or Educational Filler**:
  - BHF is designed for professional penetration testers and bug bounty hunters. Never display tutorial, explanatory, or introductory copy.
  - Remove redundant explanations of tool capabilities (e.g., do not say *"Fast passive subdomain enumeration tool querying multiple online public sources"* in card headers).
  - Do not show redundant status labels like *"Installed"* (having the tool ready is an invariant). Show only actionable status pills (`RUNNING`, `COMPLETED`, `FAILED`, `STOPPED`).
- **Zero Emojis**: Strictly prohibited across all UI elements, status badges, buttons, logs, code comments, and documentation. Use clean SVG icons (from `lucide-react`) and monospace text labels instead.

---

## 4. Offensive Security and Traffic Safety

- **Bug Bounty Value First**: Every feature, tool integration, or setting must directly contribute to high-signal reconnaissance or vulnerability identification. Features that generate superficial vanity metrics without actionable recon value are rejected.
- **Network Traffic and WAF Discipline**:
  - Never generate unthrottled request floods that saturate target bandwidth, degrade availability, or trigger rapid IP/WAF bans.
  - Enforce sane rate limits (`-rl`, `-rate`), concurrency caps (`-c`), and timeout limits (`-timeout`) on all active scanners (Katana, Ffuf, Nuclei, Httpx, Naabu).
  - Tools must default to stealthy, production-grade flags rather than aggressive brute-force defaults.
- **Universal Scope Guard**:
  - Every discovered hostname, IP, and URL must pass through `tools.IsInScope()` before insertion into SQLite.
  - Crawler and spidering tools (Katana, Gau) must enforce domain boundary flags (`-fs dn`) to prevent spidering third-party sites (e.g., CDNs, social links, OAuth providers).
- **Graceful Cancellation with Zero Data Loss**:
  - When an operator halts a scan (`STOP`), the subprocess must terminate cleanly.
  - The framework must parse all stdout/stderr streams and temporary result files accumulated up to the termination point, saving all discovered assets and findings into SQLite. Never discard collected intelligence upon cancellation.

---

## 5. Technology Stack and Architecture

- **Backend**: Statically typed, high-concurrency Go (Golang). Pure Go SQLite in WAL mode via `modernc.org/sqlite` (zero CGO dependency).
- **Single-Binary Standalone**: The Go binary embeds the built React frontend (`web/embed.go` using `embed.FS`) to produce a single self-contained executable.
- **Real-Time Streaming**: Native WebSockets (`/api/ws/scans/{id}/logs`) stream live terminal logs line-by-line and broadcast scan lifecycle updates to connected clients.
- **AI Agent Interoperability**: Built-in Model Context Protocol (MCP) server supporting stdio mode (`--mcp`) and HTTP endpoint (`POST /api/mcp`) for seamless integration with external AI agent workflows.

---

## 6. Git and Version Control Standards

- **Conventional Commits**: Commit messages must follow the Conventional Commits specification:
  - `feat`: A new end-user feature or tool integration (e.g., `feat(tools): add katana crawler integration`)
  - `fix`: A bug fix (e.g., `fix(parser): prevent panic on malformed nuclei jsonl output`)
  - `refactor`: Code restructuring without functional changes (e.g., `refactor(database): extract asset statistics calculation`)
  - `test`: Adding or updating test suites (e.g., `test(tools): add unit tests for subfinder output parsing`)
  - `docs`: Documentation updates (e.g., `docs(rules): formalize code quality standards`)
  - `chore`: Maintenance, dependencies, or tooling updates (e.g., `chore(docker): update base debian image`)
  - `perf`: Performance optimizations (e.g., `perf(database): add index on asset host and target_id`)
- **Atomic Commits**: Keep commits scoped to a single logical change. Do not mix refactors, feature additions, and unrelated fixes in a single commit.
- **Git Hygiene**: Never commit runtime database files (`bhf.db*`), compiled binaries (`bin/`), Node dependencies (`node_modules`), Python bytecode (`__pycache__`, `*.pyc`), or temporary files.

---

## 7. Python Dependency and Tool Isolation

- **Virtual Environment Requirement**: Any Python-based tool (such as `theHarvester`) must run inside an isolated virtual environment using `uv tool install` or `python -m venv`.
- **Zero Global Dependency Conflicts**: The Go server invokes Python tools via their isolated virtual environment binary paths (e.g., `/root/.local/bin/theHarvester` or `uv run`). No dependencies may be installed into global system Python site-packages.

---

## 8. Methodical Debugging Doctrine (Zero Guesswork)

- **Never Guess**: When an error, crash, or unexpected behavior occurs, never guess or speculate on the cause.
- **Root-Cause Investigation Protocol**:
  1. **Examine Process Logs**: Review the exact executed command line, stdout, stderr, and exit code.
  2. **Inspect Database State**: Query the SQLite database directly to verify table schema, inserted rows, and foreign key constraints.
  3. **Reproduce with Minimal Payload**: Isolate the failing component by reproducing the issue with a targeted test case or minimal input.
  4. **Trace Execution Path**: Follow the exact code path from router through handler, engine, repository, and models.
  5. **Apply Verified Fix**: Implement the targeted fix, run existing and new test suites, and confirm the root cause is resolved without side effects.

---

## 9. Verification Before Expansion

- **No Premature Feature Creep**: Do not proceed to adding new tools or features until all existing components are thoroughly tested, verified, and stable.
- **Regression Prevention**: Every new capability or bug fix must be accompanied by appropriate unit or integration tests to guarantee long-term stability.
