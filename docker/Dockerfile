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
FROM golang:alpine AS tools-builder
RUN apk add --no-cache git

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOTOOLCHAIN=auto

# Install Go-based reconnaissance tools as static binaries
RUN go install -v github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest && \
    go install -v github.com/projectdiscovery/httpx/cmd/httpx@latest && \
    go install -v github.com/projectdiscovery/katana/cmd/katana@latest && \
    go install -v github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest && \
    go install -v github.com/tomnomnom/assetfinder@latest && \
    go install -v github.com/lc/gau/v2/cmd/gau@latest && \
    go install -v github.com/ffuf/ffuf/v2@latest

# ==============================================================================
# Stage 3: Build the Unified BHF Go Server Binary
# ==============================================================================
FROM golang:alpine AS backend-builder
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
FROM alpine:3.20 AS runner
RUN apk add --no-cache ca-certificates bind-tools curl bash

WORKDIR /app

# Copy security tool binaries
COPY --from=tools-builder /go/bin/* /usr/local/bin/

# Copy server binary
COPY --from=backend-builder /app/bhf-server /app/bhf-server

# Directory for SQLite persistent database and scans
RUN mkdir -p /app/data

EXPOSE 8080
VOLUME ["/app/data"]

ENV BHF_PORT=8080
ENV BHF_HOST=0.0.0.0
ENV BHF_DATA_DIR=/app/data

ENTRYPOINT ["/app/bhf-server"]
