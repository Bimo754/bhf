# BHF — Bug Hunting & Security Assessment Platform

An extensible, evidence-driven security assessment and penetration testing automation platform built with Go, React, and embedded SQLite.

---

## Quick Start

### Docker (Recommended)
Runs all pre-compiled tools and execution engines in a single isolated ~350 MB container:

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

## Architectural Layers

1. **Reconnaissance & Discovery**:
   - Web: Subdomain enumeration, DNS wildcard filtering, HTTP probing, technology fingerprinting, URL archive extraction, and parameter discovery (`subfinder`, `assetfinder`, `theharvester`, `dnsx`, `httpx`, `tlsx`, `katana`, `gau`, `uro`, `jsluice`, `ffuf`, `arjun`).
   - Network: Subnet scanning, host identification, port/service enumeration, and service fingerprinting (`naabu`).
2. **Interactive Application Intelligence**:
   - Isolated browser exploration, DOM state traversal, form and API identification.
   - HTTP traffic interception, transaction categorization (authentication, registration, upload, search, admin, API), and parameter graph mapping.
3. **Intelligent Security Assessment**:
   - Hybrid AI planning + deterministic verification engines.
   - Evidence-driven testing with reproducible proof-of-concept evidence (exact HTTP transactions and `curl` commands), confidence levels, and zero speculative alert noise.
4. **Network & Infrastructure Security**:
   - Network service audit, Active Directory domain/identity mapping, privilege escalation analysis, and controlled validation.

---

## Core Engineering Principles

- **Reliability Over Quantity, Evidence Over Assumptions, Verification Over Speculation**: BHF rejects scanner-driven noise. Every reported finding must possess an explicit confidence level, verified reproducibility, and documented security impact.
- **Pure Go Backend**: Fast, compiled, single-binary distribution with embedded pure-Go SQLite in WAL mode (`modernc.org/sqlite`). Zero CGO compilation dependency.
- **100% Operator Control**: Scans and assessments execute strictly on demand with live WebSocket log streaming and graceful cancellation with zero data loss.
- **Strict Modularity (SRP & DRY)**: Single Responsibility Principle enforced across all files (1 file = 1 responsibility). Single source of truth for scope, scoring, and schema.
- **Developer-Grade UI**: Clean GitHub Dark aesthetic (`#0d1117`) with live search highlighting, instant 1-click execution, and quick asset clipboard exports. Zero AI neon glow, zero emojis.
- **Model Context Protocol (MCP)**: Native JSON-RPC 2.0 interface supporting stdio and HTTP for seamless interoperability with autonomous AI agents.

---

## Documentation

- [System Architecture](ARCHITECTURE.md): Complete system blueprint, major architectural layers, evidence lifecycle, and data models.
- [Engineering Rules](RULES.md): Mandatory architectural constraints, UI design standards, evidence requirements, and debugging rules.
- [Testing Guidelines](TESTING.md): Test architecture, deterministic verification patterns, and pre-commit checklist.
- [Development Workflow](WORKFLOW.md): Assessment methodology stages, feature lifecycle, and session starter prompt.
- [Changelog](CHANGELOG.md): Progress tracking and release history.
- [Personal Tracker](personal.md): Operator scratchpad for logging active bugs, feature roadmap, and assessment notes.
