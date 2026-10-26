# BHF - Bug Hunting Framework

A fast, lightweight, and modular bug bounty reconnaissance framework built with Go, React, and embedded SQLite.

---

## Quick Start

### Docker (Recommended)
Runs all pre-compiled recon tools in a single isolated approximately 350 MB container:

```bash
docker compose up -d --build
```

Access the web interface at **http://localhost:8080**.

### Standalone Binary
Build frontend assets and compile into a single static binary:

```bash
make build
./bin/bhf-server
```

---

## Reconnaissance Pipeline

1. **Subdomain and OSINT Enumeration**: `subfinder`, `assetfinder`, `theharvester`, `dnsx`
2. **Port and Network Scanning**: `naabu`
3. **Live Host Probing**: `httpx`
4. **Crawling and URL Discovery**: `katana`, `gau`
5. **Content Discovery and Fuzzing**: `ffuf`
6. **Vulnerability Assessment**: `nuclei`

---

## Core Engineering Principles

- **Pure Go Backend**: Fast, compiled, single-binary distribution with embedded pure-Go SQLite in WAL mode (`modernc.org/sqlite`).
- **100% Operator Control**: Tools execute strictly on demand with live WebSocket log streaming and graceful cancellation with zero data loss.
- **Strict Modularity**: Single Responsibility Principle enforced across all files (1 file = 1 responsibility).
- **Strict Scope Guard**: Automatic out-of-scope filtering (`tools.IsInScope()`) prevents third-party domain leakage.
- **Developer-Grade UI**: Clean GitHub dark aesthetic with live search highlighting, instant 1-click execution, and quick asset clipboard exports. Zero AI neon glow, zero emojis.
- **Model Context Protocol (MCP)**: Native JSON-RPC 2.0 interface for interoperability with AI agents.

---

## Documentation

- [System Architecture](ARCHITECTURE.md): Complete system blueprint, execution pipeline, and data models.
- [Engineering Rules](RULES.md): Mandatory architectural constraints, UI design standards, and debugging rules.
- [Testing Guidelines](TESTING.md): Test architecture, table-driven test patterns, and pre-commit checklist.
- [Development Workflow](WORKFLOW.md): Feature addition lifecycle, dependency isolation, and session starter prompt.
- [Changelog](CHANGELOG.md): Progress tracking and release history.
