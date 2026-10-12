.PHONY: all build build-web build-server run dev-server dev-web docker-build docker-up docker-down clean

all: build

build-web:
	@echo "==> Building frontend assets..."
	@cd web && npm run build

build-server:
	@echo "==> Compiling bhf-server binary..."
	@go build -o bin/bhf-server ./cmd/bhf-server

build: build-web build-server
	@echo "==> Build complete! Output at bin/bhf-server"

run: build
	@./bin/bhf-server

dev-server:
	@go run ./cmd/bhf-server

dev-web:
	@cd web && npm run dev

docker-build:
	@docker compose build

docker-up:
	@docker compose up -d

docker-down:
	@docker compose down

clean:
	@rm -rf bin/ data/bhf.db* web/dist
