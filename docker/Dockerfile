# ==============================================================================
# Stage 1: Build the React Frontend
# ==============================================================================
FROM node:22-alpine AS frontend-builder
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ ./
RUN npm run build

# ==============================================================================
# Stage 2: Install Pre-compiled Security Tools
# ==============================================================================
FROM golang:bookworm AS tools-builder
RUN apt-get update && apt-get install -y --no-install-recommends git libpcap-dev && rm -rf /var/lib/apt/lists/*

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOTOOLCHAIN=auto

# Install Go-based reconnaissance tools as static binaries
RUN go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest && \
    go install -v github.com/projectdiscovery/httpx/cmd/httpx@latest && \
    go install -v github.com/projectdiscovery/katana/cmd/katana@latest && \
    go install -v github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest && \
    go install -v github.com/projectdiscovery/dnsx/cmd/dnsx@latest && \
    go install -v github.com/projectdiscovery/naabu/v2/cmd/naabu@latest && \
    go install -v github.com/tomnomnom/assetfinder@latest && \
    go install -v github.com/lc/gau/v2/cmd/gau@latest && \
    go install -v github.com/ffuf/ffuf/v2@latest && \
    go install -v github.com/projectdiscovery/alterx/cmd/alterx@latest && \
    go install -v github.com/pentestmonkey/subzy@latest && \
    go install -v github.com/projectdiscovery/asnmap/cmd/asnmap@latest && \
    go install -v github.com/projectdiscovery/tlsx/cmd/tlsx@latest && \
    go install -v github.com/bishopfox/jsluice/cmd/jsluice@latest && \
    go install -v github.com/devploit/nomore403@latest && \
    go install -v github.com/sa7mon/s3scanner@latest

# ==============================================================================
# Stage 3: Build the Unified BHF Go Server Binary
# ==============================================================================
FROM golang:bookworm AS backend-builder
WORKDIR /app

ENV CGO_ENABLED=0
ENV GOTOOLCHAIN=auto

COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY cmd/ cmd/
COPY internal/ internal/
COPY web/embed.go web/
COPY --from=frontend-builder /web/dist web/dist

# Compile static Go server binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bhf-server ./cmd/bhf-server

# ==============================================================================
# Stage 4: Minimal Final Runtime Image (~350MB instead of 60GB!)
# ==============================================================================
FROM debian:bookworm-slim AS runner

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    curl \
    git \
    dnsutils \
    libpcap0.8 \
    musl \
    gcc \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy uv (fast Python package runner)
COPY --from=ghcr.io/astral-sh/uv:latest /uv /usr/local/bin/uv

# Install Python tools inside isolated virtual environments
RUN uv tool install --python 3.14 git+https://github.com/laramies/theHarvester.git && \
    uv tool install --python 3.14 wafw00f && \
    uv tool install --python 3.14 uro && \
    uv tool install --python 3.14 arjun && \
    uv tool install --python 3.14 corsy

# Copy security tool binaries
COPY --from=tools-builder /go/bin/* /usr/local/bin/

# Copy server binary
COPY --from=backend-builder /app/bhf-server /app/bhf-server

# Directory for SQLite persistent database and scans
RUN mkdir -p /app/data

EXPOSE 8080
VOLUME ["/app/data"]

ENV PATH="/root/.local/bin:/usr/local/bin:$PATH"
ENV BHF_PORT=8080
ENV BHF_HOST=0.0.0.0
ENV BHF_DATA_DIR=/app/data

ENTRYPOINT ["/app/bhf-server"]
