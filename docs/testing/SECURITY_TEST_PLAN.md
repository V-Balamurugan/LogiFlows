# LogiFlows — Application Security Test Plan

**Document Reference**: `docs/testing/SECURITY_TEST_PLAN.md`  
**Version**: 1.0.0  
**Status**: APPROVED  

---

## 1. Security Architecture Overview

LogiFlows is an enterprise multi-tenant logistics orchestration system. Its security posture relies on:
1. **Cryptographic Identity**: Bcrypt (cost 12) for passwords; HMAC-SHA256 for JWT access tokens; 256-bit cryptographically random tokens stored as SHA-256 hashes for refresh sessions.
2. **Context-Driven Tenant Isolation**: Strict isolation enforced by `middleware.TenantContext` preventing Insecure Direct Object References (IDOR).
3. **Hierarchical RBAC**: Explicit role validation (`RequireRole`) enforcing `PLATFORM_ADMIN`, `TENANT_ADMIN`, `TENANT_OPERATOR`, and `VIEWER` permission boundaries.
4. **Defense in Depth**: Panic recovery, standardized error envelopes, secret exclusion from JSON models, and immutable audit logs.

---

## 2. Threat Modeling & Security Test Matrix

| Threat Category | Attack Vector / Scenario | Target Component | Expected Defense | Test Case ID |
| :--- | :--- | :--- | :--- | :---: |
| **Authentication Bypass** | Request protected endpoint without `Authorization` header | `middleware.Auth` | Return `401 Unauthorized` with `UNAUTHORIZED` error code | `TC-P1-AUT-015` |
| **JWT Algorithm Confusion** | Craft JWT with `alg: "none"` or asymmetric public key spoof | `internal/auth/token.go` | Reject token with `401 Unauthorized` (`INVALID_TOKEN_SIGNATURE`) | `TC-P1-AUT-014` |
| **Expired Token Replay** | Submit expired JWT access token | `middleware.Auth` | Return `401 Unauthorized` with `TOKEN_EXPIRED` error code | `TC-P1-AUT-013` |
| **Token Theft & Replay** | Replay an already rotated/revoked refresh token | `internal/auth/service.go` | Detect breach, revoke entire token family for user, log `REFRESH_TOKEN_REUSE_DETECTED` | `TC-P1-AUT-017` |
| **Credential Leakage** | Inspect `users` API responses for `password_hash` | `internal/users/model.go` | `PasswordHash` is tagged with `json:"-"` and never exposed | `TC-P1-AUT-008` |
| **Cross-Tenant IDOR** | User from Tenant Alpha attempts to access Tenant Beta resource by changing UUID | `middleware.TenantContext` | Return `403 Forbidden` (`CROSS_TENANT_ACCESS_DENIED`) | `TC-P1-TNT-006` |
| **Privilege Escalation** | `TENANT_OPERATOR` attempts to invite new team member or update tenant settings | `middleware.RequireRole` | Return `403 Forbidden` (`INSUFFICIENT_PERMISSIONS`) | `TC-P1-MEM-002` |
| **SQL Injection** | Submit SQL payload in registration name, email, or tenant slug | Gin Handlers & Repositories | Parameterized SQL queries via `pgx` prevent statement tampering | `TC-P1-SEC-001` |
| **XSS Injection** | Submit HTML `<script>` tags in full name or tenant metadata | API Handlers | Inputs safely stored; responses rendered as JSON with proper escaping | `TC-P1-SEC-002` |
| **Audit Evasion** | Perform sensitive action (register, login, tenant create, member invite) | `internal/audit` | Every critical action generates immutable row in `audit_logs` | `TC-P1-AUD-001` |
| **Cross-Tenant Branch IDOR** | Tenant Alpha user attempts to view/modify Tenant Beta branch | `middleware.TenantContext` | Return `403 Forbidden` (`CROSS_TENANT_ACCESS_DENIED`) | `TC-P2-BRN-003` |
| **Foreign Branch Hijacking** | Tenant Alpha assigns employee to Tenant Beta branch ID | `internal/employees/service.go` | Reject cross-tenant branch association with `400 Bad Request` | `TC-P2-EMP-002` |
| **Cross-Tenant Vehicle IDOR** | Tenant Alpha user attempts to view or assign Tenant Beta vehicle | `internal/vehicles/service.go` | Return `403 Forbidden` (`CROSS_TENANT_ACCESS_DENIED`) | `TC-P2-VEH-004` |
| **Viewer Role Resource Mutation** | `VIEWER` attempts to create/modify branch, employee, or vehicle | `middleware.RequireRole` | Return `403 Forbidden` (`INSUFFICIENT_PERMISSIONS`) | `TC-P2-ORG-001` |
| **Double-Assignment Conflict** | Concurrently assigning already assigned vehicle or driver | PostgreSQL Partial Unique Indexes | Enforce database-level uniqueness, return `409 Conflict` | `TC-P2-VEH-003` |

---

## 3. Detailed Security Attack Vectors & Test Procedures

### 3.1 Token Algorithm Confusion (`alg: "none"`)
* **Objective**: Ensure the backend parser rejects unsigned tokens or tokens specifying `alg: "none"`.
* **Procedure**:
  1. Base64-encode header `{"alg":"none","typ":"JWT"}`.
  2. Base64-encode payload `{"sub":"<uuid>","email":"admin@logiflows.test"}`.
  3. Concatenate `<header>.<payload>.` (empty signature).
  4. Submit to `GET /api/v1/auth/me`.
* **Assertion**: API returns HTTP 401.

### 3.2 Cross-Tenant Boundary (Multi-Tenant Isolation)
* **Objective**: Prevent cross-tenant horizontal privilege escalation.
* **Procedure**:
  1. Register User A (`owner-a@logiflows.test`) and create Tenant A (`alpha-corp`).
  2. Register User B (`owner-b@logiflows.test`) and create Tenant B (`beta-corp`).
  3. User B makes a request to `GET /api/v1/tenants/<tenant-a-id>` and `GET /api/v1/tenants/<tenant-a-id>/members`.
* **Assertion**: API returns HTTP 403 Forbidden with error code `CROSS_TENANT_ACCESS_DENIED`.

### 3.3 Refresh Token Single-Use Rotation & Breach Detection
* **Objective**: Verify RFC 6749 single-use rotation and token family revocation upon breach.
* **Procedure**:
  1. User logs in, receives Access Token $AT_1$ and Refresh Token $RT_1$.
  2. User refreshes session using $RT_1$; receives $AT_2$ and $RT_2$. $RT_1$ is marked `is_revoked = true`.
  3. An attacker re-submits $RT_1$.
* **Assertion**:
  * System flags reuse.
  * System revokes $RT_2$ immediately.
  * Subsequent attempts with $RT_2$ return 401.
  * Security event `REFRESH_TOKEN_REUSE_DETECTED` is recorded in `audit_logs`.
