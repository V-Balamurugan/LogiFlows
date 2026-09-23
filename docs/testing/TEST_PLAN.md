# LogiFlows — Master Test Plan

**Document Reference**: `docs/testing/TEST_PLAN.md`  
**Version**: 1.0.0  
**Project**: LogiFlows — Intelligent End-to-End Logistics Coordination and Delivery Management System  
**Author**: Senior QA Automation Engineer & Software Test Architect  
**Status**: APPROVED  

---

## 1. Introduction & Objectives

### 1.1 Purpose
This Master Test Plan defines the comprehensive quality assurance and testing framework for the LogiFlows project. The primary directive is to guarantee that every architectural vertical slice is rigorously verified through automated testing, and that **whenever a new phase or module is developed, all previously completed features continue to operate without regression**.

### 1.2 Core Principles
1. **Evidence-Based Quality**: No feature is considered complete until verified by actual test execution. Never generate fake results, fake screenshots, or false claims.
2. **Layered Testing Pyramid**: Implement isolated unit tests, multi-service integration tests, API contract tests, security vector assessments, and full end-to-end business workflows.
3. **Module-by-Module Progression**: Maintain clear boundaries between completed phases (Phase 0, Phase 1), future planned phases (Phase 2, Phase 3), and supporting microservices (FastAPI AI service, React web frontend, Flutter mobile client).
4. **Zero-Regression Standard**: New code additions, database schema migrations, and configuration updates must automatically execute the regression suite.

---

## 2. Scope & Target Stack

### 2.1 Technology Stack Under Test
* **Backend**: Go (Gin HTTP framework, `pgxpool`, `go-redis/v9`, Goose database migrations, structured `log/slog`).
* **Databases & Caches**: PostgreSQL 16 + PostGIS 3.4 (`postgis/postgis:16-3.4`), Redis 7 (`redis:7-alpine`).
* **AI Service**: Python 3.12, FastAPI, Pydantic, Starlette TestClient.
* **Web Frontend**: React 19.2.8, TypeScript, Vite 8.3.0, Tailwind CSS, Lucide React.
* **Mobile Client**: Flutter 3.x / Dart 3.x (Material 3).
* **Containerization**: Docker Compose (`logiflows-postgres`, `logiflows-redis`).

### 2.2 In-Scope Modules
* **Phase 0 (Foundation)**:
  * Application initialization, environment variable validation, and graceful shutdown.
  * Health (`/api/v1/health`) and readiness (`/api/v1/readiness`) probes.
  * PostgreSQL + PostGIS connectivity and Redis cache ping/set/get/del.
  * Goose migration up/down execution and constraint checks.
  * Middleware pipeline: `RequestID`, `StructuredLogger`, `Recovery`, and `CORS`.
* **Phase 1 (Identity, Authentication, Authorization & Multi-Tenancy)**:
  * User registration with validation boundaries and email normalization.
  * Password hashing security (bcrypt work factor 12) and credential concealment.
  * User login and JWT access token issuance (HMAC-SHA256).
  * Opaque refresh tokens with single-use rotation, database SHA-256 digests, and breach detection.
  * Session revocation via `/api/v1/auth/logout`.
  * User profile retrieval via `/api/v1/auth/me`.
  * Tenant creation, tenant metadata updates (`PATCH /api/v1/tenants/:id`), and member invitations.
  * Context-driven multi-tenant isolation middleware (`TenantContext`).
  * Role-Based Access Control matrix (`PLATFORM_ADMIN`, `TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`).
  * Immutable database audit logging (`audit_logs`).
* **AI Service**: Delay risk prediction API and model readiness.
* **Web Frontend**: Build compilation, static linting, and authentication UI workflows.
* **End-to-End Business Scenarios**: Multi-tenant onboarding, member management, and session lifecycle.
* **Regression Testing**: Automated Phase 0 & Phase 1 verification suite.

### 2.3 Out-of-Scope (Planned / Future Phases)
* **Phase 2 (Organization & Resources)**: Branches, Employees, Vehicles, and Resource Assignments are not yet implemented in the codebase and are classified as `PLANNED / NOT IMPLEMENTED`.

---

## 3. Testing Pyramid & Methodology

```text
                  ▲
                 / \
                /E2E\             Level 4: End-to-End Business Workflows
               /-----\            (Multi-Tenant Lifecycle, Invitation, RBAC)
              /  API  \
             /Security \          Level 3: API Contracts & Security Vectors
            /-----------\         (JWT Tampering, Alg Confusion, IDOR, SQLi/XSS)
           / Integration \
          /   Database    \       Level 2: Multi-Component Integration
         /-----------------\      (PostgreSQL Pool, Redis, Goose Migrations)
        /    Unit Tests     \
       /---------------------\    Level 1: Isolated Unit Tests
                                  (Bcrypt, Token Issuance, Roles Matrix, Config)
```

### 3.1 Level 1: Unit Testing
* **Focus**: Individual business rules, validation functions, and utilities tested without external dependencies.
* **Key Targets**:
  * Bcrypt password hashing (`internal/auth/password.go`).
  * Password complexity validation rules (>= 8 chars, upper, lower, digit, symbol).
  * JWT claims signing and expiry extraction (`internal/auth/token.go`).
  * Role-Based Access Control permission matrix (`internal/authorization/roles.go`).
  * Configuration validation and fail-safe defaults (`internal/config/config.go`).
  * Context extraction helpers (`internal/contextutil`).
  * User JSON model password hash omission (`internal/users/model.go`).

### 3.2 Level 2: Integration Testing
* **Focus**: Inter-component interaction and persistence against live Docker infrastructure (`logiflows-postgres` and `logiflows-redis`).
* **Key Targets**:
  * Live PostgreSQL pool acquisition, `SELECT 1`, and PostGIS geometry capability.
  * Live Redis connection, key eviction, TTL expiration, and PING probe.
  * Goose migration execution: fresh database apply (`RunUp`) and single-step rollback & reapply (`RunDown`).
  * Domain repository CRUD operations and relational foreign key cascades.

### 3.3 Level 3: API & Security Testing
* **Focus**: HTTP request/response validation, status code contracts, header propagation, and attack vector resistance.
* **Key Targets**:
  * Positive and negative payload validation for all public and authenticated endpoints.
  * JWT tampering: forged signatures, expired tokens, and `alg: "none"` algorithm confusion.
  * Cross-tenant resource manipulation (IDOR) rejection (`403 Forbidden: CROSS_TENANT_ACCESS_DENIED`).
  * Role permission boundary enforcement (`403 Forbidden: INSUFFICIENT_PERMISSIONS`).
  * Input sanitization against SQL injection payloads and Cross-Site Scripting (XSS) markup.
  * Token refresh single-use rotation and token family breach revocation.

### 3.4 Level 4: End-to-End Business Workflows
* **Focus**: Complete user journeys spanning multiple modules and state changes.
* **Key Target**: `backend/tests/integration/e2e_workflow_test.go`
  * Owner Registration -> Login -> Tenant Provisioning -> Member Invitation -> Operator Onboarding -> Permission Check -> Cross-Tenant Rejection -> Token Refresh -> Logout.

### 3.5 Level 5: Automated Regression Testing
* **Focus**: Immutable test suite executed on every pull request and before any future phase merge.
* **Key Target**: `backend/tests/regression/phase1_regression_test.go`
  * Guarantees zero regressions in Phase 0 foundation and Phase 1 identity/security when Phase 2 or Phase 3 are introduced.

---

## 4. Test Prioritization & Defect Classification

### 4.1 Priority Levels
* **P0 (Critical / Blocker)**: Core security boundaries, authentication, password security, tenant isolation, database migrations, and data corruption risks. Test failures block release immediately.
* **P1 (High)**: Core business APIs, CRUD operations, input validation, role permissions, and error responses. Must be resolved before module sign-off.
* **P2 (Medium)**: Edge case validation, non-critical query filters, pagination, and logging format.
* **P3 (Low)**: Minor formatting, cosmetic log messages, and non-functional enhancements.

### 4.2 Severity Definitions
* **Critical**: Security breach possible, cross-tenant data leak, application crash/panic, or database corruption.
* **Major**: Primary API endpoint failure with no workaround.
* **Moderate**: Specific input validation failure or unexpected error code where a workaround exists.
* **Minor**: Typographical or documentation inconsistency.

---

## 5. Test Environment & Prerequisites

### 5.1 Infrastructure Services
```bash
# Start local infrastructure
docker compose up -d postgres redis

# Verify health status
docker compose ps
```

### 5.2 Environment Variables (`.env` or test injection)
* `APP_ENV=test`
* `APP_PORT=8080`
* `DB_HOST=localhost`
* `DB_PORT=5432`
* `DB_USER=postgres`
* `DB_PASSWORD=postgres`
* `DB_NAME=logiflows_dev`
* `REDIS_HOST=localhost`
* `REDIS_PORT=6379`
* `JWT_SECRET=super-secret-logiflows-test-jwt-key-256-bit-minimum!`
* `JWT_ACCESS_EXPIRY=15m`
* `JWT_REFRESH_EXPIRY=168h`

---

## 6. Test Deliverables & Reporting

1. **Testing Audit**: `docs/testing/TESTING_AUDIT.md` (Baseline audit of code, toolchain, and coverage gaps).
2. **Test Data Strategy**: `docs/testing/TEST_DATA_STRATEGY.md` (Fixture isolation and cleanup guidelines).
3. **Security Test Plan**: `docs/testing/SECURITY_TEST_PLAN.md` (Detailed attack matrix and assertions).
4. **Regression Test Plan**: `docs/testing/REGRESSION_TEST_PLAN.md` (Backward compatibility protocols).
5. **Phase Test Catalogs**: `TEST_CASES_PHASE_0.md`, `TEST_CASES_PHASE_1.md`, `TEST_CASES_PHASE_2.md`.
6. **Defect Tracking**: `docs/testing/BUG_REPORT.md` (Template and active defect register).
7. **Test Execution Report**: `docs/testing/TEST_EXECUTION_REPORT.md` (Empirical evidence and metrics).
