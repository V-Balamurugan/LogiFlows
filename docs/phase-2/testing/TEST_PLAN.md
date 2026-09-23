# LogiFlows Phase 2 Master Test Plan

## 1. Document Overview
This document specifies the testing strategy, scope, environment requirements, execution criteria, and quality standards for **Phase 2: Multi-Branch Logistics, Workforce Management, Fleet Assets, and Resource Scheduling** of the LogiFlows logistics management platform.

---

## 2. Test Objectives
1. **Vertical Slice Verification**: Ensure every Phase 2 module (Branches, Employees, Vehicles, Resource Assignments) works seamlessly from the PostgreSQL database layer up through Go API endpoints, React Web UI, and Flutter Mobile client.
2. **Tenant & Branch Isolation**: Enforce zero cross-tenant leakage. Confirm that tenants and branches cannot read or mutate data belonging to other organizations.
3. **Role-Based Access Control (RBAC)**: Validate that platform admins, tenant admins, branch managers, drivers, and warehouse staff can only execute operations permitted by their role. Specifically, ensure read-only (`VIEWER`) users cannot mutate resources.
4. **Data Integrity & Concurrency**: Verify that spatial coordinate boundaries, coverage radii, employee codes, unique registration numbers, and driver-vehicle non-overlapping assignments are enforced at the database level.
5. **Backwards Compatibility**: Confirm 100% regression-free operation of Phase 0 Foundation and Phase 1 Identity & Multitenancy subsystems.

---

## 3. Test Scope

### 3.1 In-Scope Components
| Component | Scope Description |
|-----------|-------------------|
| **PostgreSQL Database** | Migrations `00004_create_branches.sql`, `00005_create_employees.sql`, `00006_create_vehicles_and_assignments.sql`, PostGIS geography columns, spatial indexes, unique constraints, and partial indexes. |
| **Go Backend Core** | `internal/branches`, `internal/employees`, `internal/vehicles`, `internal/server/router.go`, Gin handlers, DTO validation, middleware authorization. |
| **React + TypeScript Frontend** | `frontend/src/pages/Branches.tsx`, `frontend/src/pages/Employees.tsx`, `frontend/src/pages/Vehicles.tsx`, `frontend/src/services/api.ts`, responsive layouts, modal forms, status badges. |
| **Flutter Mobile Client** | `mobile/lib/models/resource_models.dart`, `mobile/lib/services/resource_api_client.dart`, `mobile/lib/screens/branch_screen.dart`, `mobile/lib/screens/vehicle_screen.dart`, navigation integration in `mobile/lib/main.dart`. |
| **API Contract & Swagger UI** | OpenAPI 3.0 specification in `backend/docs/swagger.json`, `/swagger/index.html` runtime validation. |
| **Security & Isolation** | IDOR resistance, SQL injection prevention via parameterized SQL, token expiration, tamper-resistant headers. |

### 3.2 Out-of-Scope (Deferred to Phase 3)
- Real-time GPS telematics websocket streaming.
- Automated route optimization heuristics and ML parcel bin packing.
- Stripe / Payment gateway billing hooks.

---

## 4. Test Environment Architecture
- **Operating System**: Windows 11 / Linux CI
- **Database Engine**: PostgreSQL 16 with PostGIS extension (Docker container `logiflows-postgres` on port 5432)
- **Cache / Token Store**: Redis 7.2 (Docker container `logiflows-redis` on port 6379)
- **Backend Runtime**: Go 1.23+ (`cmd/api/main.go` on port 8080)
- **Frontend Runtime**: Node.js v20+, Vite 5, React 18, TypeScript 5.5 (dev server on port 5173)
- **Test Database**: `logiflows_test` / `logiflows_dev` with isolated schema generation per test run

---

## 5. Test Types & Methodologies

### 5.1 Backend Unit & Repository Tests
- Unit testing domain models, validation logic, coordinate range enforcement, payload limit checks, and DTO parsing.
- Repository layer testing using real transactional database rollbacks and mock stores.

### 5.2 End-to-End API Integration Tests
- Spinning up test HTTP servers using `httptest.NewServer` or Gin test engine.
- Making actual HTTP requests with bearer tokens, testing happy paths, validation errors, and permission failures.

### 5.3 Security & Multi-Tenant Isolation Tests
- Cross-tenant injection: Tenant A authenticates and attempts to read or mutate Tenant B's branches, employees, or vehicles.
- RBAC privilege escalation: Viewer or driver attempts to create a branch or assign vehicles.
- Concurrency testing: Simultaneously assigning two drivers to the same vehicle to verify PostgreSQL partial unique index conflict rejection (`409 Conflict`).

### 5.4 Frontend Component & Integration Tests
- Vitest + React Testing Library testing rendering, empty states, loading indicators, and API service invocation.
- Typecheck (`tsc -b`) and ESLint validation (`npm run lint`).
- Production build validation (`npm run build`).

### 5.5 Mobile Model & API Client Tests
- Flutter/Dart unit tests validating JSON serialization/deserialization, query parameter construction, and MockClient HTTP responses.

---

## 6. Entry and Exit Criteria

### 6.1 Entry Criteria
- Migrations 00001 through 00006 applied without errors.
- Go backend builds cleanly (`go build ./...`).
- Frontend dependencies installed and type definitions synced.

### 6.2 Exit Criteria (Pass Definition)
- 100% pass rate on all Go unit and integration tests.
- Zero `go vet` or `gofmt` warnings.
- 100% pass rate on React frontend component tests.
- Zero TypeScript compiler errors during `npm run build`.
- 100% pass rate on regression test suite verifying Phase 0 and Phase 1 endpoints.
- Documented test results with actual outputs.
