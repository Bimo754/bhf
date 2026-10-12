# BHF — Bug Hunting Framework

A fast, lightweight, and modular bug bounty reconnaissance framework built with Go, React, and SQLite.

---

## ⚡ Quick Start

### Docker (Recommended)
Runs all pre-compiled recon tools in a single isolated ~350 MB container:

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

## 🛠️ Reconnaissance Pipeline

1. **Subdomain Enumeration**: `subfinder`, `assetfinder`
2. **Live Host Probing**: `httpx`
3. **Crawling & URL Discovery**: `katana`, `gau`
4. **Content Discovery & Fuzzing**: `ffuf`
5. **Vulnerability Assessment**: `nuclei`

---

## 📋 Core Principles

- **Pure Go Backend**: Fast, compiled, single-binary distribution with embedded pure-Go SQLite (WAL mode).
- **100% Operator Control**: Tools execute strictly on demand with live WebSocket log streaming.
- **Strict Modularity**: 1 file = 1 responsibility across all directories.
- **Strict Scope Guard**: Automatic out-of-scope filtering ensures zero third-party domain leakage.
- **Developer-Grade UI**: Clean GitHub dark aesthetic with live search highlighting, instant 1-click execution, and quick asset clipboard exports.

For full architectural constraints and design standards, see [rules.md](rules.md).
