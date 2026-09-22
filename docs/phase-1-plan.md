# LogiFlows Phase 1 Master Plan: Identity, Multi-Tenancy & Security Foundation

## 1. Executive Summary & Objective
LogiFlows is an intelligent end-to-end logistics coordination and delivery management system designed for multi-tenant B2B operations. Independent e-commerce merchants and shipping partners integrate via LogiFlows to orchestrate shipments, track drivers, and manage deliveries.

**Phase 1 Objective**:
Establish a secure, enterprise-grade foundation for **User Identity, Authentication, Authorization (RBAC), Multi-Tenant Isolation, and Security Audit Logging**. No shipment management, driver custody, or route optimization logic is deployed until this boundary is validated.

---

## 2. SDLC & Engineering Methodology
Development strictly follows the vertical-slice lifecycle:
```text
Database Migration → Domain Model & Schemas → Repository Layer → Service Layer → API Handlers → Backend Testing → Web Frontend Integration → Mobile Foundation → Security Audit & Test Report → Module Complete
```

Key engineering principles:
- **Zero Trust Multi-Tenancy**: Tenant IDs from request bodies or parameters are never trusted without server-side verification of active membership.
- **Defense in Depth**: Password complexity validation, bcrypt cost 12, timing attack mitigation, HMAC-SHA256 JWT validation, opaque refresh token hashing, and tamper-proof audit trails.
- **Fail-Safe Defaults**: Accounts default to non-admin; roles default to minimal privileges; database transactions roll back atomically on any failure.

---

## 3. Module Breakdown & Deliverables

| Module | Scope & Objectives | Key Deliverables | Status |
| :--- | :--- | :--- | :--- |
| **Module 1: Users & Identity Database** | Goose migrations for `users` and `refresh_tokens`, case-insensitive unique email index, foreign keys, rollback testing. | `00001_enable_extensions.sql`, `00002_create_identity_and_tenancy.sql`, `00003_add_refresh_tokens_and_user_verification.sql` | **MODULE COMPLETE** |
| **Module 2: User Domain Model & Schemas** | Domain struct `users.User` with `PasswordHash` excluded from JSON, DTOs (`RegisterRequest`, `LoginRequest`, `UserSummary`), status flags. | `backend/internal/users/model.go`, `model_test.go`, `repository.go` | **MODULE COMPLETE** |
| **Module 3: Password Security** | Password complexity validator (min 8 chars, mixed case, numbers, symbols), bcrypt hashing (cost 12), constant-time dummy verification. | `backend/internal/auth/password.go`, `password_test.go` | **MODULE COMPLETE** |
| **Module 4: Authentication & Token Lifecycle** | Register (atomic user + tenant transaction), Login (JWT + opaque refresh token), Me, Refresh (single-use rotation + breach detection), Logout. | `backend/internal/auth/service.go`, `token.go`, `handler.go` | **MODULE COMPLETE** |
| **Module 5: Tenant Database** | `tenants` table with unique slug index, status check constraint (`ACTIVE`, `SUSPENDED`, `DEACTIVATED`), contact email. | Migration 00002, `backend/internal/tenants/model.go`, `repository.go` | **MODULE COMPLETE** |
| **Module 6: Tenant Memberships & RBAC** | `tenant_memberships` table, unique `(tenant_id, user_id)` constraint, roles (`PLATFORM_ADMIN`, `TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`). | `backend/internal/memberships/model.go`, `repository.go`, `authorization/roles.go` | **MODULE COMPLETE** |
| **Module 7: Tenant Management APIs** | `POST /tenants`, `GET /tenants`, `GET /tenants/:id`, `PATCH /tenants/:id`, `GET /tenants/:id/members`, `POST /tenants/:id/members`. | `backend/internal/tenants/handler.go`, `service.go` | **MODULE COMPLETE** |
| **Module 8: Middleware Enforcement** | `middleware.Auth` (Bearer JWT validation), `middleware.TenantContext` (isolation enforcement), `middleware.RequireRole` (RBAC). | `backend/internal/middleware/auth.go`, `tenant.go`, `rbac.go` | **MODULE COMPLETE** |
| **Module 9: Security Audit Logging** | `audit_logs` table recording security events with non-sensitive JSONB details, IP address, user agent, and timestamp. | `backend/internal/audit/model.go`, `repository.go` | **MODULE COMPLETE** |
| **Module 10: Backend Testing Suite** | Unit tests across all internal packages; integration tests for DB constraints, auth lifecycle, token rotation, cross-tenant rejection, and RBAC. | `backend/tests/integration/...` (100% pass) | **MODULE COMPLETE** |
| **Module 11: Web Frontend Authentication** | React 19 + TypeScript: Auth Screen (Login/Register), Token & Refresh Storage, Active Tenant Switcher, Team Modal, Protected Routes. | `frontend/src/...` (Vite build clean, 0 lint errors) | **MODULE COMPLETE** |
| **Module 12: Mobile Authentication Foundation** | Flutter / Dart mobile foundation: ApiConfig, TokenStorage interface, AuthApiClient, LoginScreen, RegisterScreen, auth navigation. | `mobile/lib/...` | **MODULE COMPLETE** |
| **Module 13: Architecture & Security Documentation** | Complete documentation suite covering schema, auth flows, RBAC matrix, API contracts, security controls, and test reports. | `docs/...` | **MODULE COMPLETE** |

---

## 4. Definition of Done Checklist
- [x] All 3 Goose database migrations applied and verified against PostgreSQL 16 + PostGIS.
- [x] Migration rollback and re-apply verified in integration tests.
- [x] Case-insensitive unique email index and password hash sanitization verified.
- [x] Password complexity enforced and tested.
- [x] JWT access tokens (HMAC-SHA256) and cryptographically secure opaque refresh tokens implemented.
- [x] Single-use refresh token rotation and breach detection cascading revocation verified.
- [x] Multi-tenant isolation verified with automated cross-tenant rejection tests returning `403 Forbidden`.
- [x] RBAC matrix verified with automated privilege escalation rejection tests.
- [x] Immutable security audit log captures all critical authentication and tenancy events.
- [x] React 19 web frontend compiles cleanly (`npm run build`) and connects to auth APIs.
- [x] Mobile authentication foundation implemented in Dart/Flutter.
- [x] 100% unit and integration test pass rate.
- [x] `gofmt -l .` reports 0 files; `go vet ./...` passes with 0 warnings.
- [x] Comprehensive documentation published to `docs/`.
