# ADR-001: Phase 0 Technology Foundation & Architectural Decisions

## Status
Accepted

## Context
LogiFlows is an Intelligent End-to-End Logistics Coordination and Delivery Management System designed for multi-company logistics operations, live parcel tracking, custody transfers, dynamic ETA calculations, and predictive intelligence. Phase 0 establishes the engineering foundation for the backend, database, caching, containerization, observability, and testing.

Key architectural choices needed evaluation to ensure performance, type safety, developer velocity, and long-term maintainability.

## Decision Drivers
1. Reliability and standard library compatibility for production Go services.
2. Low latency, high throughput, and memory efficiency under concurrent request loads.
3. Geospatial querying capabilities for parcel tracking and branch coordinates.
4. Clean migration workflows with transactional safety and rollback capabilities.
5. Zero-overhead structured observability with contextual Request ID correlation.

---

## Decisions

### 1. HTTP Web Framework: Gin
- **Options Considered**: Gin vs. Fiber vs. Standard `net/http` / Chi.
- **Decision**: **Gin (`github.com/gin-gonic/gin`)**.
- **Rationale**:
  - **Standard Library Compliance**: Gin is built on Go's standard `net/http` stack (`http.ResponseWriter`, `*http.Request`). In contrast, Fiber relies on `valyala/fasthttp`, which deviates from `net/http` interfaces and can introduce subtle memory lifecycle nuances and adapter overhead with standard Go middleware.
  - **Testing Ergonomics**: Native compatibility with Go's `net/http/httptest` allows lightning-fast, reproducible unit tests without spinning up real TCP sockets.
  - **Maturity & Ecosystem**: Gin is an industry standard with extensive production verification, battle-tested performance, rich middleware ecosystem, and active maintenance.

### 2. Database & Driver: PostgreSQL 16 + PostGIS & pgx/v5
- **Options Considered**:
  - Drivers: `jackc/pgx/v5` (pgxpool) vs. `lib/pq` vs. GORM default driver.
  - Database: PostgreSQL with PostGIS vs. vanilla PostgreSQL / MySQL.
- **Decision**: **PostgreSQL 16 with PostGIS 3.4** using **`jackc/pgx/v5/pgxpool`**.
- **Rationale**:
  - **PostGIS**: Logistics systems intrinsically demand geospatial calculations (geofencing branches, vehicle distance calculations, delivery polygon routing). PostGIS is the gold standard for spatial SQL queries.
  - **`pgx/v5`**: Significantly faster than `lib/pq` (which is in maintenance mode). `pgxpool` provides production-grade connection pooling, health checks, binary protocol support, and native Go `context.Context` cancellation.

### 3. Database Migration Tool: Goose
- **Options Considered**: Goose (`pressly/goose/v3`) vs. `golang-migrate/migrate`.
- **Decision**: **Goose**.
- **Rationale**:
  - **Idiomatic SQL Migrations**: Uses standard SQL files with simple `-- +goose Up` and `-- +goose Down` annotations.
  - **Robust Failure Handling**: Unlike `golang-migrate`, which notoriously marks migrations as "dirty" upon syntax errors (requiring manual database interventions), Goose wraps migrations in explicit transactions and rolls back cleanly on failure.
  - **Go Integration**: Supports both standalone CLI execution and native Go binary embedding (`goose.SetBaseFS`), allowing migrations to run automatically at application boot if configured.

### 4. Caching & Message Infrastructure: Redis 7
- **Options Considered**: Redis 7 vs. KeyDB vs. Memcached.
- **Decision**: **Redis 7 (`redis:7-alpine`)** with **`go-redis/v9`**.
- **Rationale**:
  - In-memory key-value caching, session management, rate limiting, and Pub/Sub capabilities reserved for future real-time tracking WebSocket broadcasts.

### 5. Observability & Structured Logging: Go `log/slog`
- **Options Considered**: `log/slog` (Standard Library) vs. Uber `zap` vs. `rs/zerolog`.
- **Decision**: **`log/slog` (Standard Library, Go 1.21+)**.
- **Rationale**:
  - Built directly into the Go standard library, eliminating third-party dependency bloat.
  - Native JSON and Text handlers with high throughput and zero allocations for common paths.
  - Seamless integration with custom context handlers to automatically extract and log `request_id` across call stacks.

---

## Consequences
- **Positive**:
  - Unified Go standard library conventions throughout the stack.
  - High performance, predictable memory profile, and rock-solid testability.
  - Ready for spatial features (PostGIS) without retrofitting databases later.
- **Trade-offs**:
  - Requires developers to write pure SQL migrations (which is preferred over ORM auto-migrations for production auditability).
