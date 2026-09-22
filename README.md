# LogiFlows

> **Intelligent End-to-End Logistics Coordination and Delivery Management System**

LogiFlows is an enterprise-grade logistics platform designed to coordinate multi-company logistics operations, custody transfers, dynamic ETA estimations, real-time GPS tracking, and AI/ML-driven delivery delay predictions.

---

## Current Status: Phase 2 — Complete (Organization, Branch, Employee, and Fleet Management)

LogiFlows has successfully completed and verified **Phase 0 (Foundation)**, **Phase 1 (Identity & Multi-Tenancy)**, and **Phase 2 (Organization & Resource Management)**:
- **Phase 0: Foundation**:
  - Modular Go backend architecture with graceful shutdown lifecycle
  - PostgreSQL 16 with PostGIS geospatial extensions & Goose transactional migrations
  - Redis 7 caching and health monitoring
  - Zero-dependency structured logging (`log/slog`) with contextual Request ID tracing
  - Liveness (`/api/v1/health`) and readiness (`/api/v1/readiness`) health probes
- **Phase 1: Identity, Authentication & Multi-Tenancy**:
  - Multi-tenant architecture with strict database isolation & RBAC matrix (`PLATFORM_ADMIN`, `TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`)
  - User registration, Bcrypt hashing (cost 12), and profile management (`/api/v1/auth/me`)
  - Dual-token authentication: HMAC-SHA256 JWT access tokens & opaque refresh tokens with single-use rotation and breach detection
  - Tenant organization lifecycle: creation with slug collision resolution, metadata update, and member invitations
  - Immutable security audit logging (`audit_logs`)
  - React + TypeScript + Vite frontend dashboard with active tenant switching and authentication flows
- **Phase 2: Multi-Tenant Companies and Distribution Branches**:
  - **Company & Tenant Management**: Dedicated endpoints (`/api/v1/companies/current`, `/api/v1/companies/:id`), organization metadata editing, compliance status, and member roster management
  - **Branch & Hub Management**: PostGIS spatial point geometry (`GEOMETRY(Point, 4326)`), coverage radius, geographic distance search (`ST_Distance`), operating status transitions (`ACTIVE`, `INACTIVE`, `SUSPENDED`), tenant-scoped uniqueness, and soft deletion
  - **Foundation Resource Management**: Minimal compatible foundation for workforce and fleet assets with partial unique index conflict prevention
  - **Web Console**: Modern React 19 + TypeScript + Vite dark-mode dashboard with `CompanyProfile`, `BranchList`, `EmployeeList`, and `VehicleList` components
  - **Flutter Mobile Application**: Material 3 mobile application in `/mobile` featuring `CompanyScreen`, `BranchScreen`, typed Dart models, and bottom tab navigation
  - **Testing & Verification**: 100% automated test pass rate across backend Go tests (42 unit, 41 integration, 17 regression), frontend Vitest tests (9/9), and mobile models
  - **API Contracts & Swagger**: Complete OpenAPI 3.0 / Swagger 1.0 specification with interactive UI at `/swagger/index.html`

---

## ⚡ Quick Start (TL;DR)

### On Linux / macOS (Bash)
```bash
# 1. Clone repository
git clone https://github.com/logiflows/logiflows.git
cd logiflows

# 2. Configure environment
cp .env.example .env

# 3. Start PostgreSQL (PostGIS) and Redis
docker compose up -d postgres redis

# 4. Run backend API
cd backend && go run ./cmd/api/main.go
```

### On Windows (PowerShell)
```powershell
# 1. Clone repository
git clone https://github.com/logiflows/logiflows.git
cd logiflows

# 2. Configure environment
Copy-Item .env.example .env

# 3. Start PostgreSQL (PostGIS) and Redis
docker compose up -d postgres redis

# 4. Run backend API
cd backend
go run ./cmd/api/main.go
```

The server will automatically:
1. Connect to PostgreSQL and Redis
2. Execute pending database migrations (enabling `postgis` and `uuid-ossp`)
3. Verify PostGIS geospatial availability
4. Launch the HTTP server on **http://localhost:8080**

---

## 📋 Table of Contents
1. [Prerequisites](#-prerequisites)
2. [How to Download the Project](#-how-to-download-the-project)
3. [Environment Configuration](#-environment-configuration)
4. [Starting Infrastructure Services](#-starting-infrastructure-services)
5. [Running the Application](#-running-the-application)
6. [Verifying the API](#-verifying-the-api)
7. [Running Tests](#-running-tests)
8. [Stopping & Resetting Services](#-stopping--resetting-services)
9. [Troubleshooting Common Issues](#-troubleshooting-common-issues)
10. [Technology Stack](#-technology-stack)
11. [Project Directory Structure](#-project-directory-structure)

---

## 🛠 Prerequisites

Before running LogiFlows, verify you have the following installed:

| Tool | Recommended Version | Verification Command |
| :--- | :--- | :--- |
| **Go** | 1.24 or later | `go version` |
| **Docker** | 25.0+ / Docker Desktop | `docker --version` |
| **Docker Compose** | v2.20+ / v5.5+ | `docker compose version` |
| **Git** | 2.40+ | `git --version` |
| **Make** *(Optional)* | Any standard Make | `make --version` |

> [!NOTE]
> Make sure Docker Desktop (or the Docker daemon) is running before proceeding to infrastructure startup.

---

## 📥 How to Download the Project

### Option A: Using Git (Recommended)
```bash
git clone https://github.com/logiflows/logiflows.git
cd logiflows
```

To switch to the active development branch:
```bash
git checkout develop
# or: git checkout feature/phase-0-foundation
```

### Option B: Downloading Source Archive
1. Download the repository source ZIP archive from GitHub.
2. Extract the archive into your desired folder.
3. Open a terminal / PowerShell in the extracted root directory:
```bash
cd LogiFlows
```

---

## ⚙ Environment Configuration

LogiFlows reads configuration from environment variables or a local `.env` file at runtime.

### 1. Copy the Environment Template
#### Linux / macOS:
```bash
cp .env.example .env
```
#### Windows (PowerShell):
```powershell
Copy-Item .env.example .env
```
#### Windows (CMD):
```cmd
copy .env.example .env
```

### 2. Default Configuration Variables
The `.env.example` file comes preconfigured with safe defaults for local development:

```env
# Application
APP_NAME=logiflows-api
APP_ENV=development
APP_HOST=0.0.0.0
APP_PORT=8080
LOG_LEVEL=debug
REQUEST_ID_HEADER=X-Request-ID

# Database (PostgreSQL 16 + PostGIS)
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=postgres
DATABASE_PASSWORD=postgres_dev_password
DATABASE_NAME=logiflows_dev
DATABASE_SSLMODE=disable
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5
DATABASE_CONN_MAX_LIFETIME=15m

# Cache & Message Broker (Redis 7)
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

> [!IMPORTANT]
> The `.env` file is excluded from Git tracking via `.gitignore`. Never commit production passwords or credentials to version control.

---

## 🐳 Starting Infrastructure Services

LogiFlows requires PostgreSQL (with PostGIS) and Redis. These are fully containerized using Docker Compose.

### 1. Start Containers in Background
```bash
# Using Makefile:
make dev-up

# Or using Docker Compose directly:
docker compose up -d postgres redis
```

### 2. Verify Container Health
Wait a few seconds for the initialization health checks to complete, then inspect the container status:
```bash
# Using Makefile:
make dev-status

# Or using Docker Compose directly:
docker compose ps
```

You should see both containers running with status `(healthy)`:
```
NAME                 IMAGE                    STATUS                   PORTS
logiflows-postgres   postgis/postgis:16-3.4   Up (healthy)             0.0.0.0:5432->5432/tcp
logiflows-redis      redis:7-alpine           Up (healthy)             0.0.0.0:6379->6379/tcp
```

### 3. View Infrastructure Logs
To view live logs from PostgreSQL or Redis:
```bash
# Follow all container logs:
docker compose logs -f

# Follow only database logs:
docker compose logs -f postgres

# Follow only Redis logs:
docker compose logs -f redis
```

### 4. Verify PostGIS Extension inside PostgreSQL
You can verify the geospatial extension directly inside the database container:
```bash
docker compose exec postgres psql -U postgres -d logiflows_dev -c "SELECT PostGIS_Full_Version();"
```

---

## 🚀 Running the Application

### Method 1: Run with Go Toolchain (Development Mode)
This is the recommended workflow during active code development:

```bash
# From the project root:
make run

# Or directly using Go:
cd backend
go run ./cmd/api/main.go
```

**Successful startup logs will display:**
```
time=... level=INFO msg="Starting LogiFlows API service" service=logiflows-api env=development log_level=info
time=... level=INFO msg="Connecting to PostgreSQL database..." host=localhost port=5432 db=logiflows_dev
time=... level=INFO msg="PostgreSQL connection established successfully"
time=... level=INFO msg="Executing database migrations..."
time=... level=INFO msg="Database migrations applied successfully"
time=... level=INFO msg="PostGIS extension active" version="POSTGIS=3.4.3 ..."
time=... level=INFO msg="Connecting to Redis cache..." host=localhost port=6379 db=0
time=... level=INFO msg="Redis connection established successfully"
time=... level=INFO msg="Starting HTTP server" service=logiflows-api env=development addr=0.0.0.0:8080
```

---

### Method 2: Compile & Run Native Binary
To compile an optimized, standalone binary:

#### Linux / macOS:
```bash
# Build binary
make build
# Or: cd backend && go build -o bin/api ./cmd/api

# Run binary
./backend/bin/api
```

#### Windows (PowerShell / CMD):
```powershell
# Build binary
cd backend
go build -v -o bin/api.exe ./cmd/api

# Run binary
.\bin\api.exe
```

---

### Method 3: Build & Run in Docker
You can also build the application into a container using the multi-stage Dockerfile:

```bash
# Build the container image
docker build -t logiflows-api:latest -f backend/Dockerfile ./backend

# Run container (connected to logiflows network)
docker run --rm -p 8080:8080 \
  --network logiflows_logiflows-network \
  -e DATABASE_HOST=postgres \
  -e REDIS_HOST=redis \
  logiflows-api:latest
```

---

## 🔍 Verifying the API

Once the server is running on `http://localhost:8080`, test the endpoints using `curl` or PowerShell:

### 1. Liveness Probe: `GET /api/v1/health`
Verifies that the HTTP service is running and accepting requests.

#### Bash / cURL:
```bash
curl -i http://localhost:8080/api/v1/health
```

#### PowerShell:
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" | ConvertTo-Json
```

**Expected Response (`200 OK`):**
```json
{
  "status": "ok",
  "service": "logiflows-api",
  "version": "v1",
  "timestamp": "2026-09-20T15:13:18.730Z"
}
```

---

### 2. Readiness Probe: `GET /api/v1/readiness`
Verifies that critical dependencies (PostgreSQL with PostGIS, Redis) are healthy and reports latency in milliseconds.

#### Bash / cURL:
```bash
curl -i http://localhost:8080/api/v1/readiness
```

#### PowerShell:
```powershell
Invoke-RestMethod -Uri "http://localhost:8080/api/v1/readiness" | ConvertTo-Json -Depth 5
```

**Expected Response (`200 OK`):**
```json
{
  "status": "ready",
  "service": "logiflows-api",
  "version": "v1",
  "timestamp": "2026-09-20T15:13:25.799Z",
  "checks": {
    "database": {
      "status": "UP",
      "latency_ms": 1.61
    },
    "redis": {
      "status": "UP",
      "latency_ms": 0.56
    }
  }
}
```

---

### 3. Request Tracing with Custom Header
Send a custom `X-Request-ID` and observe that the server preserves and logs the correlation ID:

#### Bash / cURL:
```bash
curl -i -H "X-Request-ID: my-custom-trace-12345" http://localhost:8080/api/v1/health
```

**Observed Response Header:**
```http
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
X-Request-Id: my-custom-trace-12345
```

---

### 4. Structured Error Response
Request an invalid route to verify RFC-compliant uniform error format:

#### Bash / cURL:
```bash
curl -i http://localhost:8080/api/v1/nonexistent-route
```

**Expected Response (`404 Not Found`):**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "The requested resource was not found",
    "request_id": "11e1eb4d-0940-4594-b370-fca6a543398e"
  }
}
```

---

## 🧪 Running Tests

LogiFlows includes both unit tests and integration tests.

### 1. Run Unit Tests
Unit tests do not require Docker and run in milliseconds:
```bash
# Using Makefile:
make test

# Or directly:
cd backend
go test -v ./internal/...
```

### 2. Run Integration Tests
Integration tests test live PostgreSQL (PostGIS) connection and Redis cache operations:
```bash
# Ensure Docker services are running first:
docker compose up -d postgres redis

# Run integration tests:
make test-integration

# Or directly:
cd backend
go test -v ./tests/integration/...
```

### 3. Run Static Code Analysis (Vet & Formatting)
```bash
# Verify Go formatting:
cd backend && gofmt -l .

# Run static analysis:
make lint
# Or: cd backend && go vet ./...
```

---

## 🛑 Stopping & Resetting Services

### 1. Stop the Backend API
In the terminal where `go run` or `./api` is executing, press:
```
Ctrl + C
```
The server will catch `SIGINT` and initiate a graceful shutdown within 10 seconds:
```
time=... level=INFO msg="Received shutdown signal" signal=interrupt
time=... level=INFO msg="Commencing graceful service shutdown (10s timeout)..."
time=... level=INFO msg="HTTP server stopped gracefully"
time=... level=INFO msg="LogiFlows API service shutdown complete"
```

### 2. Stop Docker Infrastructure Services
```bash
# Using Makefile:
make dev-down

# Or using Docker Compose:
docker compose down
```

### 3. Reset Local Database Volume (Fresh Start)
> [!WARNING]
> This command will permanently delete all local development database and cache data.
```bash
# Using Makefile:
make dev-reset

# Or using Docker Compose:
docker compose down -v
docker compose up -d postgres redis
```

---

## 🔧 Troubleshooting Common Issues

### Issue 1: "port is already allocated" (5432, 6379, or 8080)
- **Cause**: A previous instance of PostgreSQL, Redis, or an HTTP server is already running on your host machine.
- **Fix (Windows PowerShell)**:
  ```powershell
  # Find which process is using port 5432 or 8080:
  Get-NetTCPConnection -LocalPort 5432, 8080 -ErrorAction SilentlyContinue
  # Or change the port in .env, e.g.:
  # APP_PORT=8081
  # DATABASE_PORT=5433
  ```
- **Fix (Linux / macOS)**:
  ```bash
  lsof -i :5432
  lsof -i :8080
  ```

### Issue 2: Docker daemon is not running
- **Symptom**: `Cannot connect to the Docker daemon` or `error during connect`.
- **Fix**: Launch Docker Desktop and ensure the engine status shows **Engine running** before executing `docker compose up -d`.

### Issue 3: Backend reports "dial tcp: connection refused" on startup
- **Symptom**: `Failed to connect to PostgreSQL` or `Failed to connect to Redis`.
- **Fix**: Check if containers are still starting up:
  ```bash
  docker compose ps
  ```
  Wait for both containers to display status `(healthy)`, then start the backend.

---

## 🏗 Technology Stack

| Layer | Technology | Purpose |
| :--- | :--- | :--- |
| **Core Backend** | Go 1.24+ | High-throughput concurrency, microservices |
| **HTTP Routing** | Gin (`gin-gonic/gin`) | `net/http`-compliant REST routing |
| **Database** | PostgreSQL 16 + PostGIS 3.4 | Relational datastore & geospatial calculations |
| **Driver & Pooling**| `jackc/pgx/v5` (`pgxpool`) | High-performance connection pooling & context timeouts |
| **Migrations** | Goose (`pressly/goose/v3`) | Pure SQL transactional migrations with embedded runner |
| **Cache & Real-Time**| Redis 7 (`redis:7-alpine`) | Key-value caching & telemetry Pub/Sub |
| **Observability** | Go `log/slog` | Zero-dependency structured JSON logging |
| **Containerization** | Docker & Docker Compose | Reproducible local development & production containerization |
| **CI/CD** | GitHub Actions | Automated formatting, vetting, testing, and compilation |

---

## 📁 Project Directory Structure

```
LogiFlows/
├── backend/
│   ├── cmd/api/                # Application entrypoint (main.go)
│   ├── internal/               # Private application packages
│   │   ├── config/             # Strongly typed configuration & validation
│   │   ├── database/           # PostgreSQL connection pool & PostGIS verification
│   │   ├── health/             # Liveness & Readiness probe handlers
│   │   ├── logger/             # slog structured logger & context request ID
│   │   ├── middleware/         # Request ID, structured logger, panic recovery
│   │   ├── redis/              # Redis client & ping health checks
│   │   ├── response/           # Uniform API response envelopes
│   │   └── server/             # Router setup & graceful HTTP server
│   ├── migrations/             # Goose SQL migrations & runner
│   │   ├── 00001_enable_extensions.sql
│   │   └── migrate.go
│   ├── tests/integration/      # Integration test suite against live infrastructure
│   ├── Dockerfile              # Multi-stage production container definition
│   ├── go.mod                  # Go module definition
│   └── go.sum                  # Cryptographic dependency checksums
├── docs/                       # Comprehensive documentation
│   ├── architecture/           # System architecture & component design
│   ├── api/                    # API standards & endpoint documentation
│   ├── decisions/              # Architecture Decision Records (ADR-001)
│   ├── project-journal/        # Chronological engineering execution journal
│   └── testing/                # Verification logs & test reports
├── frontend/                   # Reserved for React + Vite + TypeScript (Future Phase)
├── mobile/                     # Reserved for Flutter + Dart (Future Phase)
├── ai-service/                 # Reserved for FastAPI + Scikit-Learn (Future Phase)
├── infrastructure/             # Deployment configurations & scripts
├── .github/workflows/          # CI pipeline automation
├── docker-compose.yml          # Local development infrastructure services
├── .env.example                # Canonical environment variable template
├── Makefile                    # Developer automation targets
└── README.md                   # Project overview & onboarding guide
```

---

## 📜 Git & Branching Conventions

### Branch Strategy
- `main`: Production-ready release branch
- `develop`: Integration branch for completed vertical slices
- `feature/<name>`: Feature branch (e.g. `feature/phase-0-foundation`)

### Commit Message Standard
All commit messages must strictly follow the professional **Phase & Component Scope** standard:

```
Phase <N> - <Particular Part>: <Imperative summary of changes>
```

#### Approved Phase 0 Examples:
- `Phase 0 - Repository & Environment: Initialize directory structure and toolchain configuration`
- `Phase 0 - Configuration: Implement typed environment loading and validation`
- `Phase 0 - Backend Core: Implement Gin router, server skeleton, and uniform response envelopes`
- `Phase 0 - Observability: Implement slog structured JSON logging and Request ID middleware`
- `Phase 0 - Database & PostGIS: Configure PostgreSQL connection pool and PostGIS extension verification`
- `Phase 0 - Migrations: Configure Goose migrations with extensions migration`
- `Phase 0 - Cache: Implement Redis client, connection management, and ping check`
- `Phase 0 - Health & Readiness: Implement liveness and readiness probe endpoints`
- `Phase 0 - Infrastructure: Configure Docker Compose for PostgreSQL/PostGIS and Redis`
- `Phase 0 - Testing: Add unit tests and live Docker integration tests`
- `Phase 0 - CI & DevOps: Configure GitHub Actions CI workflow and Makefile`
- `Phase 0 - Documentation: Add architecture diagram, ADR-001, API standards, and execution journal`
- `Phase 0 - Onboarding: Provide complete download, run, and troubleshooting guide in README`

---

## 📄 License
Proprietary & Confidential — LogiFlows Engineering Team.
