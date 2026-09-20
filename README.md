# LogiFlows

> **Intelligent End-to-End Logistics Coordination and Delivery Management System**

LogiFlows is an enterprise-grade logistics platform designed to coordinate multi-company logistics operations, custody transfers, dynamic ETA estimations, real-time GPS tracking, and AI/ML-driven delivery delay predictions.

---

## Current Status: Phase 0 — Complete Project Initialization

LogiFlows is currently in **Phase 0**. Phase 0 establishes the engineering and architectural foundation for the entire platform:
- Clean modular Go backend architecture
- PostgreSQL 16 with PostGIS geospatial extension
- Redis 7 caching and event-broker foundation
- Goose SQL migration framework
- High-performance Gin HTTP router with standard `net/http` compatibility
- Zero-dependency structured JSON logging with `log/slog`
- Contextual Request ID propagation and tracing
- Graceful shutdown lifecycle
- Liveness and readiness health checks
- Unit and integration testing suites
- GitHub Actions CI workflow

---

## Technology Stack

| Layer | Technology | Purpose |
| :--- | :--- | :--- |
| **Core Backend** | Go 1.24+ | Microservices, concurrency, high-throughput APIs |
| **HTTP Routing** | Gin (`gin-gonic/gin`) | High-performance, `net/http`-compliant REST routing |
| **Database** | PostgreSQL 16 + PostGIS 3.4 | Primary relational datastore + geospatial queries |
| **Driver & Pooling**| `jackc/pgx/v5` (`pgxpool`) | Connection pooling, binary wire protocol, context timeouts |
| **Migrations** | Goose (`pressly/goose/v3`) | Pure SQL transactional migrations with embedded runner |
| **Cache & Real-Time**| Redis 7 (`redis:7-alpine`) | Key-value caching, future tracking Pub/Sub telemetry |
| **Observability** | Go `log/slog` | Structured JSON logging with contextual Request IDs |
| **Containerization** | Docker & Docker Compose | Containerized local infrastructure & reproducible builds |
| **CI/CD** | GitHub Actions | Automated formatting, vetting, testing, and compilation |

---

## Directory Structure

```
LogiFlows/
├── backend/
│   ├── cmd/api/                # Application entrypoint (main.go)
│   ├── internal/               # Private application packages
│   │   ├── config/             # Environment configuration & validation
│   │   ├── database/           # PostgreSQL pool & PostGIS verification
│   │   ├── health/             # Liveness & Readiness probe handlers
│   │   ├── logger/             # slog structured logger & context request ID
│   │   ├── middleware/         # Request ID, structured logger, panic recovery
│   │   ├── redis/              # Redis client & ping health checks
│   │   ├── response/           # Uniform API response envelopes
│   │   └── server/             # Router setup & graceful HTTP server
│   ├── migrations/             # Goose SQL migrations & runner
│   ├── tests/integration/      # Integration test suite against live infrastructure
│   ├── Dockerfile              # Multi-stage production container definition
│   ├── go.mod                  # Go module definition
│   └── go.sum                  # Cryptographic dependency checksums
├── docs/                       # Project documentation
│   ├── architecture/           # System architecture & component design
│   ├── api/                    # API standards & endpoint documentation
│   ├── decisions/              # Architecture Decision Records (ADRs)
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

## Getting Started

### Prerequisites
- **Go**: 1.24 or later
- **Docker & Docker Compose**: Docker 25+ / Docker Desktop
- **Git**: 2.40+

### 1. Clone & Configure Environment
```bash
git clone https://github.com/logiflows/logiflows.git
cd logiflows

# Copy environment configuration
cp .env.example .env
```

### 2. Start Infrastructure Services
Start PostgreSQL (with PostGIS) and Redis in background containers:
```bash
make dev-up
# Or: docker compose up -d postgres redis
```

Check service status:
```bash
make dev-status
# Or: docker compose ps
```

### 3. Run the Backend API
```bash
make run
# Or: cd backend && go run ./cmd/api/main.go
```

The server will automatically:
1. Connect to PostgreSQL and Redis
2. Apply pending Goose migrations (enabling `postgis` and `uuid-ossp`)
3. Verify PostGIS geospatial availability
4. Launch the HTTP server on port `8080`

---

## Foundation Endpoints

### 1. Liveness Probe: `GET /api/v1/health`
Verifies process responsiveness.
```bash
curl -i http://localhost:8080/api/v1/health
```
**Response (200 OK):**
```json
{
  "status": "ok",
  "service": "logiflows-api",
  "version": "v1",
  "timestamp": "2026-09-20T20:30:00Z"
}
```

### 2. Readiness Probe: `GET /api/v1/readiness`
Verifies PostgreSQL (PostGIS) and Redis dependencies.
```bash
curl -i http://localhost:8080/api/v1/readiness
```
**Response (200 OK):**
```json
{
  "status": "ready",
  "service": "logiflows-api",
  "version": "v1",
  "timestamp": "2026-09-20T20:30:00Z",
  "checks": {
    "database": {
      "status": "UP",
      "latency_ms": 1.15
    },
    "redis": {
      "status": "UP",
      "latency_ms": 0.85
    }
  }
}
```

---

## Testing & Quality Assurance

### Run Unit Tests
```bash
make test
# Or: cd backend && go test -v ./internal/...
```

### Run Integration Tests (Requires Docker containers active)
```bash
make test-integration
# Or: cd backend && go test -v ./tests/integration/...
```

### Run Static Analysis
```bash
make lint
# Or: cd backend && go vet ./...
```

### Build Binary
```bash
make build
# Or: cd backend && go build -v -o bin/api.exe ./cmd/api
```

---

## Git & Branching Conventions
- `main`: Production-ready release branch
- `develop`: Integration branch for completed vertical slices
- `feature/<name>`: Work branches for specific modules (e.g. `feature/phase-0-foundation`)
- Commit style: [Conventional Commits](https://www.conventionalcommits.org/)
  - `feat(scope): ...`
  - `fix(scope): ...`
  - `chore(scope): ...`
  - `test(scope): ...`
  - `docs(scope): ...`

---

## License
Proprietary & Confidential — LogiFlows Engineering Team.
