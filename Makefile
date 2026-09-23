.PHONY: help dev-up dev-down dev-logs dev-status dev-reset run test test-integration test-all lint build migrate-up migrate-down migrate-status

help: ## Display available commands
	@echo "LogiFlows Development Automation"
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'

dev-up: ## Start PostgreSQL (PostGIS) and Redis in background
	docker compose up -d postgres redis

dev-down: ## Stop local infrastructure containers
	docker compose down

dev-logs: ## Follow infrastructure container logs
	docker compose logs -f

dev-status: ## Check health and status of Docker services
	docker compose ps

dev-reset: ## Reset local database volume (WARNING: deletes all local dev data)
	@echo "WARNING: Resetting local Docker volumes..."
	docker compose down -v
	docker compose up -d postgres redis

run: ## Run the Go backend API locally
	cd backend && go run ./cmd/api/main.go

test: ## Run backend unit tests
	cd backend && go test -v ./internal/...

test-integration: ## Run integration tests against running Docker services
	cd backend && go test -v ./tests/integration/...

test-all: ## Run all tests (unit + integration)
	cd backend && go test -v ./...

lint: ## Run Go static analysis (vet)
	cd backend && go vet ./...

build: ## Build backend binary
	cd backend && go build -v -o bin/api.exe ./cmd/api

run-all: ## Run the entire project (backend, ai-service, frontend, mobile)
	powershell -NoProfile -ExecutionPolicy Bypass -File ./run-all.ps1

stop-all: ## Stop all running LogiFlows services
	powershell -NoProfile -ExecutionPolicy Bypass -File ./stop-all.ps1
