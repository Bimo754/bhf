# BHF (Bug Hunting Framework)

A next-generation, high-performance, and modular bug bounty reconnaissance framework built from the ground up to replace bloated, unstable legacy frameworks.

---

## ⚡ Why BHF? (Comparison with Legacy Frameworks like Ars0n)

| Feature | Ars0n Framework v2 | BHF (Bug Hunting Framework) |
|---|---|---|
| **Architecture** | 50+ Docker containers (~60 GB disk footprint) | **1 Single Container** (~350 MB) or **1 Static Binary** (~19 MB) |
| **Backend & Concurrency** | Python scripts + Bash glue + monolithic API | **Pure Go (Golang)** + Goroutines + cancellation contexts |
| **Frontend UI** | 7,260-line monolithic `App.js` with 150+ `useState` | **React 18 + Vite + Tailwind CSS + Lucide Icons** (Modular, 1 file 1 responsibility) |
| **Tool Control** | Blind automatic background loops | **100% User-controlled execution** (User must click to run) |
| **Live Logs** | Unreliable or delayed log viewing | **Real-time WebSocket streaming terminal** with autoscroll, stop, & filter |
| **Database** | Heavy PostgreSQL container setup | **Embedded SQLite in WAL mode** (zero extra containers, blazing fast) |
| **Code Modularity** | Multiple responsibilities tangled in single files | **Strict Single Responsibility Principle**: 1 file = 1 functionality |

---

## 📁 Repository Structure

BHF adheres to a strict directory layout where every file has exactly one clear responsibility:

```
bhf/
├── cmd/
│   └── bhf-server/
│       └── main.go                 # Application bootstrap & lifecycle
├── internal/
│   ├── config/                     # Configuration structs & environment loading
│   │   ├── config.go               # Config struct
│   │   └── loader.go               # Environment loader & defaults
│   ├── models/                     # Core domain models
│   │   ├── target.go               # Target program model
│   │   ├── scan.go                 # Tool execution session model
│   │   ├── asset.go                # Discovered asset model
│   │   └── log.go                  # Streamed log line model
│   ├── database/                   # SQLite persistence layer (WAL mode)
│   │   ├── connection.go           # High-performance DB connection pool
│   │   ├── schema.go               # Table DDL & migrations
│   │   ├── target_repo.go          # Target CRUD operations
│   │   ├── scan_repo.go            # Scan execution history & status
│   │   ├── asset_repo.go           # Asset deduplication & pagination
│   │   └── log_repo.go             # Live log append & historical retrieval
│   ├── engine/                     # Concurrency & execution engine
│   │   ├── stream.go               # WebSocket streaming packet model
│   │   ├── hub.go                  # WebSocket client pub/sub hub
│   │   ├── runner.go               # Process runner with context cancellation
│   │   └── dispatcher.go           # Active execution lifecycle & user cancel
│   ├── tools/                      # Tool wrappers (1 file per tool)
│   │   ├── interface.go            # Common Tool interface
│   │   ├── registry.go             # Tool registry & discovery
│   │   ├── subfinder.go            # Subfinder wrapper & JSON parser
│   │   ├── httpx.go                # Httpx prober wrapper & JSON parser
│   │   ├── assetfinder.go          # Assetfinder wrapper & line parser
│   │   ├── gau.go                  # GAU wrapper & URL parser
│   │   ├── katana.go               # Katana crawler & JSON parser
│   │   ├── nuclei.go               # Nuclei template scanner & JSON parser
│   │   └── ffuf.go                 # FFuf fuzzer wrapper & line parser
│   └── api/                        # HTTP REST API & WebSockets
│       ├── response.go             # Standard JSON response envelopes
│       ├── router.go               # Chi router, CORS, and SPA routing
│       ├── target_handler.go       # Target endpoints
│       ├── scan_handler.go         # Tool execution & cancellation endpoints
│       ├── asset_handler.go        # Paginated asset queries
│       ├── tool_handler.go         # Tool catalog & status
│       └── ws_handler.go           # WebSocket log streaming endpoint
├── web/                            # Modern UI (Vite + React + Tailwind CSS)
│   ├── src/
│   │   ├── types/                  # TypeScript interface definitions
│   │   ├── services/               # API & WebSocket client services
│   │   ├── components/             # Reusable UI widgets (Terminal, Badges, Tables)
│   │   ├── views/                  # Dedicated views (Tools, Assets, Console)
│   │   ├── App.tsx                 # View state coordinator (< 180 lines)
│   │   └── main.tsx                # React DOM mount
│   ├── embed.go                    # Go embed.FS embedding dist into binary
│   └── package.json
├── docker/
│   └── Dockerfile                  # Single multi-stage slim Docker image
├── docker-compose.yml              # Single-service Docker orchestration
├── Makefile                        # Convenient development commands
└── README.md
```

---

## 🚀 Quick Start

### Option 1: Docker (Recommended for complete toolset)
Launching BHF via Docker runs all pre-compiled tools in a single isolated ~350MB container:

```bash
docker compose up -d --build
```
Access the dashboard at **http://localhost:8080**.

### Option 2: Local Native Build
Build both frontend and server into a single static binary:

```bash
make build
./bin/bhf-server
```

### Option 3: Development Mode
Run backend and frontend independently for hot-reloading:

```bash
# Terminal 1: Backend
make dev-server

# Terminal 2: Frontend with Vite HMR
make dev-web
```
- Backend API & WebSocket: `http://localhost:8080`
- Frontend with hot reload: `http://localhost:3000`

---

## 🎯 How Tool Execution Works

1. **Target Selection**: Create or select a scope program (e.g. `example.com`).
2. **User Initiation**: Click **Run Tool** on any tool card (e.g., Subfinder, Httpx, Nuclei). Tools do **not** run automatically without your trigger.
3. **Live Streaming**: The UI connects to the WebSocket stream and displays stdout/stderr output line by line in real time inside the interactive console.
4. **Immediate Parsing**: As tools complete, results are parsed and upserted directly into the SQLite database.
5. **Asset Exploration**: Switch to **Attack Surface Assets** to view, search, and paginate discovered hosts, status codes, ports, and technologies.
