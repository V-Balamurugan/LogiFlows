# LogiFlows Engineering Project Journal — Phase 1: Identity & Multi-Tenancy

## 1. Phase Metadata
- **Project**: LogiFlows (Intelligent End-to-End Logistics Coordination and Delivery Management System)
- **Phase**: Phase 1 — Identity & Multi-Tenancy Foundation
- **Methodology**: Agile Scrum / Domain-Driven Design / Clean Architecture
- **Branch**: `feature/phase-1-identity-multitenancy`
- **Status**: **MODULE COMPLETE**
- **Date**: 2026-09-21

---

## 2. Business Reason & Objective
LogiFlows is a multi-tenant B2B logistics coordination platform where independent e-commerce merchants and logistics partners integrate through secure APIs.
To prevent catastrophic cross-company data leakage and establish secure operator accountability:
1. Every partner company must be isolated into a distinct **Tenant**.
2. All operational APIs must enforce server-side validation of active **Tenant Membership**.
3. Cryptographic password hashing and signed JWT authentication must protect user accounts.
4. Role-Based Access Control (RBAC) must restrict operational vs administrative functions.

---

## 3. Scope & Boundaries

### 3.1 In-Scope
- User identity registration and authentication with email normalization and uniqueness.
- Strong password complexity policy (min 8 chars, uppercase, lowercase, digit, symbol) with `bcrypt` (cost 12).
- Signed HMAC-SHA256 JWT access tokens with claim validation and expiration.
- Atomic B2B onboarding: Company (Tenant) creation + User creation + `TENANT_ADMIN` assignment in one database transaction.
- Multi-tenant data isolation: `RequireTenantContext` middleware rejecting cross-tenant queries with `403 Forbidden` (`CROSS_TENANT_ACCESS_DENIED`).
- Centralized RBAC model (`PLATFORM_ADMIN`, `TENANT_ADMIN`, `TENANT_OPERATOR`, `VIEWER`).
- Append-only security audit trail (`audit_logs`) capturing registrations, logins, member additions, and failures.
- React 19 + TypeScript authentication UI (Login, Register, Tenant Switcher, Team Management Modal, Protected Routes).

### 3.2 Out-of-Scope (Deferred to Later Phases)
- OAuth2 / Social Login (Google / GitHub).
- SMS OTP driver verification (Phase 3 driver custody app).
- B2B API Key generation and HMAC webhook signing (Phase 2 partner operations).
- Courier partner fleet and branch assignment (Phase 2).

---

## 4. Database Architecture & Schema Design

### 4.1 Schema Definition (`backend/migrations/00002_create_identity_and_tenancy.sql`)
The migration establishes four normalized tables with foreign keys, cascaded deletions, and integrity constraints:

```sql
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(50),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    is_platform_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_lower ON users (LOWER(email));
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users (is_active);

-- Tenants table (Companies)
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'SUSPENDED', 'DEACTIVATED')),
    contact_email VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_slug ON tenants (slug);
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants (status);

-- Tenant Memberships table (RBAC associations)
CREATE TABLE IF NOT EXISTS tenant_memberships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL CHECK (role IN ('PLATFORM_ADMIN', 'TENANT_ADMIN', 'TENANT_OPERATOR', 'VIEWER')),
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INVITED', 'DEACTIVATED')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_user UNIQUE (tenant_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_memberships_user_id ON tenant_memberships (user_id);
CREATE INDEX IF NOT EXISTS idx_memberships_tenant_id ON tenant_memberships (tenant_id);

-- Audit Logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    ip_address VARCHAR(45),
    user_agent TEXT,
    status VARCHAR(50) NOT NULL,
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_created ON audit_logs (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created ON audit_logs (user_id, created_at DESC);
```

---

## 5. Security Decisions & Implementation Details

1. **Password Security**:
   - Validated against 5 complexity rules: min 8 characters, at least 1 uppercase, 1 lowercase, 1 digit, 1 special symbol.
   - Hashed with `bcrypt` (work factor 12).
   - Password hashes explicitly tagged `json:"-"` on domain structs to ensure zero exposure.
2. **JWT Signing & Algorithm Confusion Prevention**:
   - `TokenService` cryptographically signs with `HMAC-SHA256`.
   - `ValidateAccessToken` explicitly verifies `t.Method.(*jwt.SigningMethodHMAC)` to prevent algorithm confusion attacks.
   - Startup validation in `config.go` enforces minimum 32-character secret length.
3. **Multi-Tenant Isolation Strategy**:
   - The server never trusts tenant IDs supplied in request bodies.
   - `TenantContext` middleware checks that `user_id` holds an active membership in the target `tenant_id`.
   - Requests targeting unauthorized tenants immediately return `403 Forbidden` (`CROSS_TENANT_ACCESS_DENIED`).
4. **Timing Attack Mitigation**:
   - On failed logins where an email does not exist, a constant-time bcrypt verification is executed against a dummy hash to neutralize user enumeration attacks.

---

## 6. API Specifications (`/api/v1`)

| Method | Endpoint | Access | Description |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/auth/register` | Public | Registers user + company in an atomic transaction; assigns `TENANT_ADMIN`. |
| `POST` | `/api/v1/auth/login` | Public | Validates credentials, records audit log, returns access token + memberships. |
| `GET` | `/api/v1/auth/me` | Bearer Token | Retrieves authenticated user profile and authorized company memberships. |
| `POST` | `/api/v1/auth/logout` | Bearer Token | Records logout audit event; client purges cached token. |
| `POST` | `/api/v1/tenants` | Bearer Token | Creates additional tenant company. |
| `GET` | `/api/v1/tenants` | Bearer Token | Lists companies available to caller. |
| `GET` | `/api/v1/tenants/:tenant_id` | Bearer + Tenant Context | Retrieves tenant metadata (isolated). |
| `GET` | `/api/v1/tenants/:tenant_id/members` | Bearer + Tenant Context | Lists team members (Requires `TENANT_ADMIN`, `TENANT_OPERATOR`, or `VIEWER`). |
| `POST` | `/api/v1/tenants/:tenant_id/members` | Bearer + Tenant Context | Adds member to company (Requires `TENANT_ADMIN`). |

---

## 7. Automated Test Evidence

### 7.1 Static Analysis & Linting
```text
$ gofmt -l .
(Clean - 0 warnings)

$ go vet ./...
(Clean - 0 warnings)
```

### 7.2 Unit Tests (`go test -v ./internal/...`)
- `internal/auth`: 8/8 PASS (`TestValidatePassword_Rules`, `TestHashAndComparePassword`, `TestTokenService_GenerateAndValidate`, `TestTokenService_ExpiredToken`, `TestTokenService_WrongSecret`).
- `internal/authorization`: 2/2 PASS (`TestRolePermissions` across all role hierarchies, `TestIsValidRole`).
- `internal/config`: 7/7 PASS (`TestConfig_Load_Default`, port validation, pool constraints, JWT secret length validation).
- `internal/health`: 3/3 PASS (Liveness, Readiness healthy, Readiness degraded DB).
- `internal/logger`: 2/2 PASS (JSON format, request ID correlation).
- `internal/middleware`: 3/3 PASS (UUID generation, header preservation, panic recovery).

### 7.3 Integration & Security Tests (`go test -v ./tests/integration/...`)
```text
=== RUN   TestAuthAPI_Register_Login_Me_Lifecycle
--- PASS: TestAuthAPI_Register_Login_Me_Lifecycle (1.79s)
=== RUN   TestAuthAPI_ExpiredToken_Rejected
--- PASS: TestAuthAPI_ExpiredToken_Rejected (0.03s)
=== RUN   TestMigrations_RunUp_Success
--- PASS: TestMigrations_RunUp_Success (0.05s)
=== RUN   TestPostgreSQL_ConnectionAndPostGIS
--- PASS: TestPostgreSQL_ConnectionAndPostGIS (0.13s)
=== RUN   TestRedis_ConnectionAndPing
--- PASS: TestRedis_ConnectionAndPing (0.02s)
=== RUN   TestRepositories_CRUD_And_Constraints
--- PASS: TestRepositories_CRUD_And_Constraints (0.06s)
=== RUN   TestSecurity_CrossTenantAccess_Forbidden
--- PASS: TestSecurity_CrossTenantAccess_Forbidden (0.86s)
=== RUN   TestSecurity_RBAC_RolePermissionEnforcement
--- PASS: TestSecurity_RBAC_RolePermissionEnforcement (2.46s)
=== RUN   TestSecurity_InvalidTenantID_Rejected
--- PASS: TestSecurity_InvalidTenantID_Rejected (0.96s)
PASS
ok  	github.com/logiflows/logiflows/backend/tests/integration	6.994s
```

### 7.4 Frontend Compilation & Packaging (`npm run build`)
```text
> tsc -b && vite build
vite v8.3.0 building client environment for production...
✓ 1883 modules transformed.
dist/index.html                   0.86 kB │ gzip:  0.48 kB
dist/assets/index-CyoUAJBu.css    1.73 kB │ gzip:  0.83 kB
dist/assets/index-Dy_8ePV3.js   270.85 kB │ gzip: 80.50 kB
✓ built in 1.28s
```

---

## 8. Files Created and Modified

### Created Files
- `backend/migrations/00002_create_identity_and_tenancy.sql`
- `backend/internal/auth/password.go` & `password_test.go`
- `backend/internal/auth/token.go` & `token_test.go`
- `backend/internal/auth/dto.go`
- `backend/internal/auth/service.go`
- `backend/internal/auth/handler.go`
- `backend/internal/users/model.go`
- `backend/internal/users/repository.go`
- `backend/internal/tenants/model.go`
- `backend/internal/tenants/dto.go`
- `backend/internal/tenants/repository.go`
- `backend/internal/tenants/service.go`
- `backend/internal/tenants/handler.go`
- `backend/internal/memberships/model.go`
- `backend/internal/memberships/repository.go`
- `backend/internal/audit/model.go`
- `backend/internal/audit/repository.go`
- `backend/internal/authorization/roles.go` & `roles_test.go`
- `backend/internal/contextutil/context.go`
- `backend/internal/middleware/auth.go`
- `backend/internal/middleware/tenant.go`
- `backend/internal/middleware/rbac.go`
- `backend/tests/integration/migration_test.go`
- `backend/tests/integration/repository_test.go`
- `backend/tests/integration/auth_api_integration_test.go`
- `backend/tests/integration/tenant_security_integration_test.go`
- `frontend/src/types/auth.ts`
- `frontend/src/services/api.ts`
- `frontend/src/context/AuthContext.tsx`
- `frontend/src/components/auth/AuthScreen.tsx`
- `frontend/src/components/auth/TopNav.tsx`
- `frontend/src/components/tenants/TeamModal.tsx`
- `docs/project-journal/phase-1.md`

### Modified Files
- `backend/go.mod` (aligned to Go 1.24, added `jwt/v5` and `golang.org/x/crypto`)
- `backend/internal/config/config.go` & `config_test.go` (added `JWTConfig` and validations)
- `backend/internal/response/response.go` (added `Unauthorized`, `Forbidden`, `Conflict` helpers)
- `backend/internal/server/router.go` (wired Phase 1 routes and middlewares)
- `backend/cmd/api/main.go` (wired Phase 1 repositories, services, and handlers)
- `.env.example` & `.env` (added JWT configuration keys)
- `frontend/src/App.tsx` (wrapped with `AuthProvider`, conditionally rendering `AuthScreen` or `Dashboard` with `TopNav`)

---

## 9. Errors Encountered & Resolved

| Error | Root Cause | Fix Applied |
| :--- | :--- | :--- |
| `gofmt -l .` reported 20 files | Windows CRLF line endings saved by Git | Converted all `.go` source files to Unix LF endings; ran `gofmt -w .`. |
| `go build` import cycle between `middleware`, `auth`, and `response` | Mutual package dependency for context getters and response helpers | Created `internal/contextutil` package for request context getters/setters; removed `middleware` import from `response/response.go`. |
| `tsc -b` verbatimModuleSyntax error | TypeScript types imported as runtime values | Replaced `import { ... }` with `import type { ... }` across all types in `api.ts`, `AuthContext.tsx`, and `TeamModal.tsx`. |
| `go: -race requires cgo` on Windows | Local Windows system lacks a configured GCC/MinGW compiler | Executed unit and integration tests directly in Go runtime; race testing remains configured in Linux CI where GCC is standard. |

---

## 10. Conclusion & Handover
Phase 1 (Identity & Multi-Tenancy) is **100% complete and fully verified**.
All acceptance criteria have been satisfied with live test evidence. The platform is ready for Phase 2: Partner Management & Operational Foundations.
