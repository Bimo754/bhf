# BHF (Bug Hunting Framework) — Engineering & Design Rules

This document defines the strict architectural constraints, design rules, and operator preferences established for the BHF project. All future code additions, modifications, and AI prompts must adhere strictly to these rules.

---

## 1. System & Environment Integrity
- **Zero Host Package Pollution**: Never run `pacman`, `apt`, or install security tools directly on the host machine. Never install global Python pip packages.
- **Single Container Deployment**: All recon tools and backend services must run inside the single, multi-stage Docker container (`docker compose up -d --build`, ~350 MB footprint) or isolated project virtual environments.
- **Never Recreate Ars0n's Architecture**: Do not create dozens of microservice containers or multi-gigabyte bloat. Keep everything unified, lightweight, and fast.

---

## 2. Technology Stack & Performance
- **Backend in Go (Golang)**: High-speed, compiled, statically typed backend. Avoid Python and shell scripts wherever Go can do the job.
- **Embedded Database**: Pure Go SQLite in WAL mode (`modernc.org/sqlite`). Zero CGO dependencies.
- **Single Binary Capability**: The compiled Go binary embeds the frontend via `embed.FS` for zero-dependency standalone execution.
- **Real-Time Streaming**: Use native WebSockets for bidirectional live log streaming and scan lifecycle notifications.

---

## 3. Strict Code Modularity (Single Responsibility Principle)
- **1 File = 1 Functionality**: Strictly enforced across both backend and frontend. It is prohibited to combine multiple distinct functionalities into a single file.
- **Structured Directories**:
  - `internal/models/`: One domain entity per file (`target.go`, `scan.go`, `asset.go`, `log.go`).
  - `internal/database/`: One repository per entity (`target_repo.go`, `scan_repo.go`, `asset_repo.go`, `log_repo.go`).
  - `internal/engine/`: Dedicated files for process runner, WebSocket hub, and scan coordinator.
  - `internal/tools/`: Exactly one file per security tool (`subfinder.go`, `httpx.go`, `katana.go`, etc.).
  - `web/src/components/`: Granular, reusable UI components (`ToolCard.tsx`, `Terminal.tsx`, `AssetTable.tsx`, `Toast.tsx`, `ScopeDropdown.tsx`).
  - `web/src/views/`: Dedicated page views (`ToolsView.tsx`, `AssetsView.tsx`, `TerminalView.tsx`).

---

## 4. UI / UX Design & Aesthetic Standards
- **Zero AI Tropes**: Strictly NO neon cyan/purple/magenta gradients, no emojis, and no AI-generated look.
- **Developer-Grade Dark Theme**: GitHub Dark / Linear dark aesthetic (`#0d1117` canvas, `#161b22` cards/panels, `#30363d` borders, `#21262d` active states).
- **Subdued, Eye-Friendly Colors**: Avoid oversaturated colors that cause eye strain.
- **Selection Visibility**: Global `::selection` must always use high-contrast styling (`#1f6feb` background with crisp `#ffffff` text). Never use low-contrast selection backgrounds that make text vanish.
- **No Native HTML Form Chrome**: Do not use raw, unstyled HTML `<select>` elements. Use custom, dark-styled popover components (e.g., `ScopeDropdown`).
- **NO FLUFF / NO UNNECESSARY EXPLANATIONS**:
  - BHF is a tool for professional security researchers and bug bounty hunters—**never include educational or tutorial text**.
  - Do NOT explain what tools or elements are (e.g., remove descriptions like *"Subdomain Enumeration — Passive OSINT... Systematically discover..."* or *"bin: subfinder"* or *"Fast passive subdomain enumeration tool..."*).
  - Keep only clean methodology category headers (`Subdomain Enumeration`, `Live Host Probing`, `Crawling & URL Discovery`, `Content Discovery`, `Vulnerability Assessment`).
  - Do NOT display redundant statuses like *"Installed"* (having the tool installed is mandatory) or *"Last run"*.

---

## 5. Execution & Operator Control
- **100% User-Initiated Execution**: Tools must NEVER run in automated loops or start without explicit operator action.
- **Instant 1-Click Run**: Clicking **Run** must immediately launch the tool with optimal pre-tuned defaults. Do NOT interrupt execution with mandatory confirmation modals.
- **Dedicated Config Button**: Provide a separate sliders/settings button (`Sliders`) next to Run for operators who wish to override CLI flags.
- **Graceful Cancellation with Zero Loss**: When an operator stops a scan, cancel the process gracefully and **always parse and save all findings discovered up to that moment** into SQLite. Never discard collected findings.
- **Toast Notifications**: Every scan completion, failure, or stop must trigger an auto-dismissing toast notification in the corner.
- **Sleek Status Indicators**: Display clean, minimal pills next to tool names (`RUNNING` pulsing blue, `COMPLETED` green, `FAILED` red, `STOPPED` zinc).

---

## 6. Live Console & Terminal
- **Horizontal Session Tabs**: Switch execution sessions using a sleek tab bar across the top of the terminal (tool name, status dot, timestamp), not a dropdown.
- **Active Command Display**: Show the exact executed command with a 1-click clipboard copy button.
- **Human-Readable Live Output**: Stream clean terminal lines, banners, and progress updates. Never dump raw JSON blobs or entire HTTP response bodies onto stdout.
- **Live Search Highlighting**: Filtering terminal output must visually highlight matching text in yellow/black (`<mark>`) and display an active match counter.
- **Isolated Scrolling**: Terminal containers must strictly isolate their scroll behavior. Log output must never hijack or scroll the outer browser window.

---

## 7. Scope Integrity & Target Scoping
- **Strict In-Scope Enforcement**: Any URL or hostname discovered whose domain does not match or is not a subdomain of the target (`target.Domain` or `*.target.Domain`) is OUT OF SCOPE.
- **Universal Scope Guard**: The backend must filter all parsed findings through `tools.IsInScope()` before inserting them into SQLite.
- **Crawler Scoping**: Tools like Katana and GAU must enforce domain scoping (`-fs dn`) to prevent spidering third-party sites (e.g., Twitter/X, Discord, Google Fonts).

---

## 8. Discovered Assets Inventory
- **Real-Time Category Counts**: All filter chips (`All`, `Live Web`, `2xx OK`, `3xx Redirect`, `4xx / 5xx`) must display live counts computed from the database.
- **Auto-Refresh**: Assets must re-fetch automatically upon tab switch or scan completion without requiring a full page refresh. Provide a manual Refresh button.
- **Hunter Clipboard Export**: Provide 1-click buttons to copy all discovered hosts (`Copy Hosts`) and live URLs (`Copy URLs`) to clipboard for CLI chaining.
- **Direct Live Browsing**: Render clickable links with external-link icons to open live targets in a new tab.
