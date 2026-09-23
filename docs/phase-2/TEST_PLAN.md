# LogiFlows — Phase 2 Master Test Plan

**Document Reference**: `docs/phase-2/TEST_PLAN.md`  
**Execution Date**: 2026-09-22  
**Phase**: Phase 2 — Multi-Tenant Companies and Distribution Branches  
**Status**: APPROVED & VERIFIED  

---

## 1. Document Overview & Objectives

This document specifies the comprehensive testing strategy, scope, environment configuration, and acceptance criteria for **Phase 2: Multi-Tenant Companies and Distribution Branches** of the LogiFlows logistics coordination platform.

### Core Quality Objectives
1. **Vertical Slice Verification**: Ensure company and branch management operate end-to-end across PostgreSQL/PostGIS, Go backend Gin routes, React web console, and Flutter mobile client.
2. **Tenant & Branch Isolation**: Guarantee 100% boundary isolation with zero cross-tenant leakage. Confirm that tenants and branches cannot read or mutate data belonging to other organizations.
3. **Role-Based Access Control (RBAC)**: Validate permissions across platform admins, company admins, dispatchers/operators, viewers, and drivers.
4. **PostGIS Spatial Integrity**: Confirm coordinates are stored using WGS 84 (`GEOMETRY(Point, 4326)`), bounds checking is enforced (-90 to +90 lat, -180 to +180 lng), and radius proximity queries compute geodesic distances correctly.
5. **Operational Status State Machine**: Enforce valid status transitions (`ACTIVE`, `INACTIVE`, `SUSPENDED`) across companies and distribution hubs.
6. **Zero Regression Guarantee**: Confirm 100% backward compatibility with Phase 0 foundation probes and Phase 1 identity/multitenancy subsystems.

---

## 2. Test Scope & Matrix

| Component | Target Artifacts | Scope Description |
| :--- | :--- | :--- |
| **PostgreSQL Database** | Migrations `00001` - `00004` | PostGIS spatial point geometry, spatial GIST index, compound unique constraint `(tenant_id, branch_code)`, foreign keys, soft deletion. |
| **Go Backend** | `internal/tenants`, `internal/branches`, `internal/server/router.go` | Route aliases (`/companies`, `/tenants`), `GET /companies/current`, `PATCH /branches/:id/status`, DTO validation, error envelopes. |
| **Security & Auth** | `internal/auth`, `internal/middleware` | HMAC-SHA256 JWT access tokens, SHA-256 refresh rotation, IDOR resistance, SQL injection prevention. |
| **React Web Console** | `CompanyProfile.tsx`, `BranchList.tsx`, `api.ts` | Organization profile editing, branch listing with PostGIS coordinate chips, branch creation dialog, status toggles. |
| **Flutter Mobile Client** | `CompanyModel`, `BranchModel`, `ResourceApiClient`, `CompanyScreen`, `BranchScreen` | Mobile Dart models, JSON serialization, mock HTTP unit tests. |
| **OpenAPI / Swagger** | `backend/docs/swagger.json`, `/swagger/index.html` | Swagger UI loading, schema definition accuracy, route documentation. |

---

## 3. Test Types & Execution Strategy

### 3.1 Backend Unit & DTO Validation Tests
* Coordinate bounding checks: Reject lat < -90 or > 90, lng < -180 or > 180.
* String length and format validation: Alphanumeric branch code format, required address fields.
* Operating status enumeration validation (`ACTIVE`, `INACTIVE`, `SUSPENDED`).

### 3.2 Database Integration Tests
* Real transactional connections using PostgreSQL 16 + PostGIS 3.4 in Docker.
* Spatial proximity validation using `ST_DWithin` and `ST_Distance`.
* Migration up/rollback idempotency verification.

### 3.3 Security & Multi-Tenant Boundary Tests
* Cross-tenant read/write prevention (IDOR attacks).
* Role privilege escalation prevention (Viewer mutations rejected with 403 Forbidden).
* Access token tamper resistance (algorithm confusion, forged signatures).

### 3.4 Frontend Component & Build Tests
* Vitest component tests verifying token storage, form validation, and API envelope parsing.
* TypeScript strict compilation (`tsc -b`).
* Vite production bundling (`npm run build`).

### 3.5 Mobile Unit Tests
* Dart model deserialization from backend JSON response envelopes.
* Mock HTTP client integration verifying Bearer token propagation and 404 error handling.

---

## 4. Entry and Exit Criteria

### 4.1 Entry Criteria
- Database container running with PostGIS extension enabled.
- All migrations 00001 through 00004 applied cleanly.
- Go backend compiles with zero build errors.
- Frontend dependencies installed and type checked.

### 4.2 Exit Criteria (Pass Definition)
- 100% pass rate on all Go unit, integration, security, and regression tests.
- Zero `go vet` or `gofmt` warnings.
- 100% pass rate on React frontend component tests (`npm test -- --run`).
- Zero TypeScript compiler errors during `npm run build`.
- Zero critical or high-severity vulnerabilities in security audit.
- Full execution evidence documented with actual test outputs.
